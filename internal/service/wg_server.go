package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/core/ip"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/core/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/network"
	"github.com/HappyLadySauce/errors"
	"k8s.io/klog/v2"
)

// WGServerSrv defines the interface for WireGuard server configuration management.
type WGServerSrv interface {
	GetServerConfig(ctx context.Context) (*wireguard.InterfaceConfig, string, string, string, error)
	UpdateServerConfig(ctx context.Context, req *v1.UpdateServerConfigRequest) error
}

type wgServerSrv struct {
	service       *service
	store         store.Factory
	configManager *wireguard.ServerConfigManager
}

// WGServerSrv if implemented, then wgServerSrv implements WGServerSrv interface.
var _ WGServerSrv = (*wgServerSrv)(nil)

func newWGServer(s *service) *wgServerSrv {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		klog.V(1).InfoS("WireGuard config not available, config manager will be nil")
		return &wgServerSrv{
			service:       s,
			store:         s.store,
			configManager: nil,
		}
	}

	wgOpts := cfg.WireGuard
	configPath := wgOpts.ServerConfigPath()
	configManager := wireguard.NewServerConfigManager(configPath, wgOpts.ApplyMethod)

	return &wgServerSrv{
		service:       s,
		store:         s.store,
		configManager: configManager,
	}
}

// GetServerConfig gets the server configuration.
// Returns: InterfaceConfig, PublicKey, ServerIP, DNS, error
func (w *wgServerSrv) GetServerConfig(ctx context.Context) (*wireguard.InterfaceConfig, string, string, string, error) {
	if w.configManager == nil {
		return nil, "", "", "", errors.WithCode(code.ErrWGConfigNotInitialized, "config manager not initialized")
	}

	// Read server config
	serverConfig, err := w.configManager.ReadServerConfig()
	if err != nil {
		return nil, "", "", "", errors.Wrap(err, "failed to read server config")
	}

	if serverConfig.Interface == nil {
		return nil, "", "", "", errors.WithCode(code.ErrWGServerConfigNotFound, "server interface config not found")
	}

	// Get server public key
	var publicKey string
	if serverConfig.Interface.PrivateKey != "" {
		publicKey, err = w.configManager.GetServerPublicKey()
		if err != nil {
			klog.V(1).InfoS("failed to get server public key", "error", err)
			// Continue without public key
		}
	}

	// Get or detect ServerIP
	cfg := config.Get()
	var serverIP string
	var dns string
	if cfg != nil && cfg.WireGuard != nil {
		serverIP = cfg.WireGuard.ServerIP
		if serverIP == "" {
			// Auto-detect if not set
			detectedIP, err := network.GetServerIP(ctx, "")
			if err != nil {
				klog.V(1).InfoS("failed to detect server IP", "error", err)
				// Continue with empty serverIP
			} else {
				serverIP = detectedIP
			}
		}
		dns = cfg.WireGuard.DNS
	}

	return serverConfig.Interface, publicKey, serverIP, dns, nil
}

// UpdateServerConfig updates the server configuration.
func (w *wgServerSrv) UpdateServerConfig(ctx context.Context, req *v1.UpdateServerConfigRequest) error {
	if w.configManager == nil {
		return errors.WithCode(code.ErrWGConfigNotInitialized, "config manager not initialized")
	}

	// Read current config
	serverConfig, err := w.configManager.ReadServerConfig()
	if err != nil {
		return errors.Wrap(err, "failed to read server config")
	}

	if serverConfig.Interface == nil {
		return errors.WithCode(code.ErrWGServerConfigNotFound, "server interface config not found")
	}

	// Save old config for comparison
	oldConfig := &wireguard.InterfaceConfig{
		PrivateKey: serverConfig.Interface.PrivateKey,
		Address:    serverConfig.Interface.Address,
		ListenPort: serverConfig.Interface.ListenPort,
		DNS:        serverConfig.Interface.DNS,
		MTU:        serverConfig.Interface.MTU,
		PreUp:      serverConfig.Interface.PreUp,
		PostUp:     serverConfig.Interface.PostUp,
		PreDown:    serverConfig.Interface.PreDown,
		PostDown:   serverConfig.Interface.PostDown,
		SaveConfig: serverConfig.Interface.SaveConfig,
	}

	// Merge updates (only update provided fields)
	if req.Address != nil {
		serverConfig.Interface.Address = *req.Address
	}
	if req.ListenPort != nil {
		serverConfig.Interface.ListenPort = *req.ListenPort
	}
	if req.PrivateKey != nil {
		// Validate private key
		if err := wireguard.ValidatePrivateKey(*req.PrivateKey); err != nil {
			return errors.Wrap(err, "invalid private key")
		}
		serverConfig.Interface.PrivateKey = *req.PrivateKey
		// Clear public key cache when private key changes
		// (ServerConfigManager will regenerate it on next GetServerPublicKey call)
	}
	if req.MTU != nil {
		serverConfig.Interface.MTU = *req.MTU
	}
	if req.PostUp != nil {
		serverConfig.Interface.PostUp = *req.PostUp
	}
	if req.PostDown != nil {
		serverConfig.Interface.PostDown = *req.PostDown
	}

	// Handle ServerIP and DNS update
	cfg := config.Get()
	var oldServerIP, oldDNS string
	if cfg != nil && cfg.WireGuard != nil {
		// Save old values for comparison
		oldServerIP = cfg.WireGuard.ServerIP
		oldDNS = cfg.WireGuard.DNS

		if req.ServerIP != nil {
			// Update ServerIP if provided
			cfg.WireGuard.ServerIP = *req.ServerIP
		} else if cfg.WireGuard.ServerIP == "" {
			// Auto-detect if not set and not provided
			detectedIP, err := network.GetServerIP(ctx, "")
			if err != nil {
				klog.V(1).InfoS("failed to detect server IP", "error", err)
				// Continue without serverIP
			} else {
				cfg.WireGuard.ServerIP = detectedIP
				klog.V(1).InfoS("auto-detected server IP", "ip", detectedIP)
			}
		}
		if req.DNS != nil {
			// Update DNS if provided (empty string means clear DNS setting)
			cfg.WireGuard.DNS = *req.DNS
		}
		// Note: Endpoint is not updated here as it's not in UpdateServerConfigRequest
		// Endpoint is only set via command line/config file
	}

	// Write updated config
	if err := w.configManager.WriteServerConfig(serverConfig); err != nil {
		return errors.Wrap(err, "failed to write server config")
	}

	// Apply config (reload WireGuard interface)
	if err := w.configManager.ApplyConfig(); err != nil {
		klog.V(1).InfoS("failed to apply server config", "error", err)
		// Continue anyway, config is written but not applied
	}

	// Detect changes in global config
	serverIPChanged := cfg != nil && cfg.WireGuard != nil && cfg.WireGuard.ServerIP != oldServerIP
	dnsChanged := cfg != nil && cfg.WireGuard != nil && cfg.WireGuard.DNS != oldDNS
	// Endpoint change detection would require adding it to UpdateServerConfigRequest
	// For now, we'll only handle ServerIP and DNS changes

	// Sync client configs if needed
	if err := w.syncClientConfigs(ctx, oldConfig, serverConfig.Interface, serverIPChanged, dnsChanged, false); err != nil {
		klog.V(1).InfoS("failed to sync client configs", "error", err)
		// Continue anyway, server config is updated
	}

	return nil
}

// syncClientConfigs synchronizes all client configurations when server config changes.
func (w *wgServerSrv) syncClientConfigs(ctx context.Context, oldConfig, newConfig *wireguard.InterfaceConfig, serverIPChanged, dnsChanged, endpointChanged bool) error {
	// Check what changed
	listenPortChanged := oldConfig.ListenPort != newConfig.ListenPort
	mtuChanged := oldConfig.MTU != newConfig.MTU
	privateKeyChanged := oldConfig.PrivateKey != newConfig.PrivateKey

	// If nothing changed that affects clients, return early
	if !listenPortChanged && !mtuChanged && !privateKeyChanged && !serverIPChanged && !dnsChanged && !endpointChanged {
		return nil
	}

	// Calculate new public key if private key changed
	var newPublicKey string
	if privateKeyChanged {
		var err error
		newPublicKey, err = wireguard.GeneratePublicKey(newConfig.PrivateKey)
		if err != nil {
			return errors.Wrap(err, "failed to generate new public key")
		}
	}

	// Get all active peers
	peers, _, err := w.store.WGPeers().ListPeers(ctx, store.WGPeerListOptions{
		Status: model.WGPeerStatusActive,
		Limit:  10000, // Get all active peers
	})
	if err != nil {
		return errors.Wrap(err, "failed to list peers")
	}

	// Get global config for Endpoint IP
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return errors.WithCode(code.ErrWGConfigNotInitialized, "WireGuard config not initialized")
	}
	wgOpts := cfg.WireGuard

	// Extract Endpoint IP from global config
	endpointIP, err := ip.ExtractIPFromEndpoint(wgOpts.Endpoint)
	if err != nil {
		klog.V(1).InfoS("failed to extract endpoint IP", "endpoint", wgOpts.Endpoint, "error", err)
		// Continue without endpoint IP update
	}

	// Update each peer's client config
	for _, peer := range peers {
		if err := w.updatePeerClientConfig(ctx, peer, newConfig, newPublicKey, endpointIP, listenPortChanged, mtuChanged, privateKeyChanged); err != nil {
			klog.V(1).InfoS("failed to update peer client config", "peerID", peer.ID, "error", err)
			// Continue with other peers
		}
	}

	// Handle Endpoint related changes (ServerIP, ListenPort, or Endpoint changed)
	if listenPortChanged || serverIPChanged || endpointChanged {
		// Update Peer endpoints in database
		if err := w.service.WGPeers().UpdatePeersEndpointForGlobalConfigChange(ctx); err != nil {
			klog.V(1).InfoS("failed to update peers endpoint for global config change", "error", err)
			// Continue anyway
		}

		// Update IP Pool endpoints in database
		if err := w.service.IPPools().UpdateIPPoolsEndpointForGlobalConfigChange(ctx); err != nil {
			klog.V(1).InfoS("failed to update IP pools endpoint for global config change", "error", err)
			// Continue anyway
		}
	}

	// Handle DNS changes
	if dnsChanged {
		// Update Peer DNS in database
		if err := w.service.WGPeers().UpdatePeersDNSForGlobalConfigChange(ctx); err != nil {
			klog.V(1).InfoS("failed to update peers DNS for global config change", "error", err)
			// Continue anyway
		}

		// Update IP Pool DNS in database
		if err := w.service.IPPools().UpdateIPPoolsDNSForGlobalConfigChange(ctx); err != nil {
			klog.V(1).InfoS("failed to update IP pools DNS for global config change", "error", err)
			// Continue anyway
		}
	}

	return nil
}

// updatePeerClientConfig updates a single peer's client configuration.
func (w *wgServerSrv) updatePeerClientConfig(ctx context.Context, peer *model.WGPeer, newConfig *wireguard.InterfaceConfig, newPublicKey, endpointIP string, listenPortChanged, mtuChanged, privateKeyChanged bool) error {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return errors.WithCode(code.ErrWGConfigNotInitialized, "WireGuard config not initialized")
	}

	wgOpts := cfg.WireGuard

	// Get server public key
	var serverPublicKey string
	if privateKeyChanged {
		serverPublicKey = newPublicKey
	} else {
		var err error
		if w.configManager != nil {
			serverPublicKey, err = w.configManager.GetServerPublicKey()
			if err != nil {
				return errors.Wrap(err, "failed to get server public key")
			}
		}
	}

	// Use DNS directly from database - it's already calculated and stored during create/update
	dns := peer.DNS

	// Use endpoint from database, but update port if ListenPort changed
	endpoint := peer.Endpoint
	if listenPortChanged {
		// Extract IP from current endpoint if it exists
		currentEndpointIP, err := ip.ExtractIPFromEndpoint(endpoint)
		if err == nil && currentEndpointIP != "" {
			// Use current endpoint IP with new port
			endpoint = fmt.Sprintf("%s:%d", currentEndpointIP, newConfig.ListenPort)
		} else {
			// Use Settings.ServerIP or detected IP
			serverIP := wgOpts.ServerIP
			if serverIP == "" {
				detectedIP, err := network.GetServerIP(ctx, "")
				if err == nil {
					serverIP = detectedIP
				}
			}
			if serverIP != "" {
				endpoint = fmt.Sprintf("%s:%d", serverIP, newConfig.ListenPort)
			} else if endpointIP != "" {
				// Fallback to extracted endpoint IP from global config
				endpoint = fmt.Sprintf("%s:%d", endpointIP, newConfig.ListenPort)
			}
		}
	}

	// Use AllowedIPs directly from database
	allowedIPs := peer.AllowedIPs
	// Only calculate default if it's still empty (shouldn't happen for new peers, but handle legacy data)
	if allowedIPs == "" {
		// Get IP pool configuration if peer has IPPoolID
		var pool *model.IPPool
		if peer.IPPoolID != "" {
			var err error
			pool, err = w.store.IPPools().GetIPPool(ctx, peer.IPPoolID)
			if err == nil && pool != nil && pool.Routes != "" {
				allowedIPs = pool.Routes
			}
		}
		// Fallback to global default if still empty
		if allowedIPs == "" {
			allowedIPs = wgOpts.DefaultAllowedIPs
		}
	}

	// Determine MTU
	mtu := 0
	if mtuChanged {
		mtu = newConfig.MTU
	} else {
		// Use existing MTU from client config if available
		// For now, we'll use the server MTU if it's set
		if newConfig.MTU > 0 {
			mtu = newConfig.MTU
		}
	}

	// Generate client config
	clientConfig := &wireguard.ClientConfig{
		PrivateKey:          peer.ClientPrivateKey,
		Address:             peer.ClientIP,
		DNS:                 dns,
		MTU:                 mtu,
		Endpoint:            endpoint,
		PublicKey:           serverPublicKey,
		AllowedIPs:          allowedIPs,
		PersistentKeepalive: peer.PersistentKeepalive,
	}

	configContent := wireguard.GenerateClientConfig(clientConfig)

	// Save to file
	userDir := wgOpts.ResolvedUserDir()
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return errors.WithCode(code.ErrWGUserDirCreateFailed, "failed to create user directory: %s", err.Error())
	}

	// Use peer ID as filename
	configPath := filepath.Join(userDir, peer.ID+".conf")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return errors.WithCode(code.ErrWGConfigWriteFailed, "failed to write client config file: %s", err.Error())
	}

	klog.V(2).InfoS("updated client config file", "peerID", peer.ID, "path", configPath, "listenPortChanged", listenPortChanged, "mtuChanged", mtuChanged, "privateKeyChanged", privateKeyChanged)
	return nil
}
