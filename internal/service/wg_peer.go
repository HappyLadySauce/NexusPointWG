package service

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/core/ip"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/core/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/NexusPointWG/pkg/options"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/network"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/snowflake"
	"github.com/HappyLadySauce/errors"
	"k8s.io/klog/v2"
)

// WGPeerSrv defines the interface for WireGuard peer business logic.
type WGPeerSrv interface {
	CreatePeer(ctx context.Context, userID, deviceName, ipPoolID, clientIP, allowedIPs, dns, endpoint, clientPrivateKey string, persistentKeepalive *int) (*model.WGPeer, error)
	GetPeer(ctx context.Context, id string) (*model.WGPeer, error)
	GetPeerByPublicKey(ctx context.Context, publicKey string) (*model.WGPeer, error)
	UpdatePeer(ctx context.Context, peer *model.WGPeer, newClientIP, newIPPoolID *string) error
	DeletePeer(ctx context.Context, id string, isHardDelete bool) error
	ListPeers(ctx context.Context, opt store.WGPeerListOptions) ([]*model.WGPeer, int64, error)
	ReleaseIP(ctx context.Context, peerID string) error
	CountPeersByUserID(ctx context.Context, userID string) (int64, error)
	UpdatePeersForIPPoolChange(ctx context.Context, poolID string, newPool *model.IPPool) error
	UpdatePeersEndpointForGlobalConfigChange(ctx context.Context) error
	UpdatePeersDNSForGlobalConfigChange(ctx context.Context) error
}

type wgPeerSrv struct {
	store         store.Factory
	configManager *wireguard.ServerConfigManager
}

// WGPeerSrv if implemented, then wgPeerSrv implements WGPeerSrv interface.
var _ WGPeerSrv = (*wgPeerSrv)(nil)

func newWGPeers(s *service) *wgPeerSrv {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		klog.V(1).InfoS("WireGuard config not available, config manager will be nil")
		return &wgPeerSrv{
			store:         s.store,
			configManager: nil,
		}
	}

	wgOpts := cfg.WireGuard
	configPath := wgOpts.ServerConfigPath()
	configManager := wireguard.NewServerConfigManager(configPath, wgOpts.ApplyMethod)

	return &wgPeerSrv{
		store:         s.store,
		configManager: configManager,
	}
}

func (w *wgPeerSrv) CreatePeer(ctx context.Context, userID, deviceName, ipPoolID, clientIP, allowedIPs, dns, endpoint, clientPrivateKey string, persistentKeepalive *int) (*model.WGPeer, error) {
	// Get default IP pool if not specified
	var pool *model.IPPool
	if ipPoolID == "" {
		pools, _, err := w.store.IPPools().ListIPPools(ctx, store.IPPoolListOptions{
			Status: model.IPPoolStatusActive,
			Limit:  1,
		})
		if err != nil || len(pools) == 0 {
			return nil, errors.WithCode(code.ErrIPPoolNotFound, "no active IP pool found")
		}
		pool = pools[0]
		ipPoolID = pool.ID
	} else {
		// Get IP pool to use its configuration
		var err error
		pool, err = w.store.IPPools().GetIPPool(ctx, ipPoolID)
		if err != nil {
			return nil, err
		}
	}

	// Use IP pool configuration if peer fields are not specified
	// Priority: Peer specified > IP Pool config > Global config
	if allowedIPs == "" && pool.Routes != "" {
		allowedIPs = pool.Routes
	}
	// Fallback to global default if still empty
	if allowedIPs == "" {
		cfg := config.Get()
		if cfg != nil && cfg.WireGuard != nil && cfg.WireGuard.DefaultAllowedIPs != "" {
			allowedIPs = cfg.WireGuard.DefaultAllowedIPs
		}
	}

	// Get server tunnel IP from server config Address
	serverTunnelIP := ""
	if w.configManager != nil {
		serverConfig, err := w.configManager.ReadServerConfig()
		if err == nil && serverConfig != nil && serverConfig.Interface != nil && serverConfig.Interface.Address != "" {
			// Extract IP from Address (e.g., "100.100.100.1/24" -> "100.100.100.1")
			extractedIP, err := ip.ExtractIPFromCIDR(serverConfig.Interface.Address)
			if err == nil {
				serverTunnelIP = extractedIP
			}
		}
	}

	// Allocate IP address
	allocator := ip.NewAllocator(w.store)
	var allocatedIP string
	if clientIP != "" {
		// Validate and use provided IP
		if err := allocator.ValidateAndAllocateIP(ctx, ipPoolID, clientIP, serverTunnelIP); err != nil {
			return nil, err
		}
		allocatedIP = clientIP
	} else {
		// Auto-allocate IP
		var err error
		allocatedIP, err = allocator.AllocateIP(ctx, ipPoolID, "", serverTunnelIP)
		if err != nil {
			return nil, err
		}
	}

	// Format IP as CIDR
	clientIPCIDR, err := ip.FormatIPAsCIDR(allocatedIP)
	if err != nil {
		return nil, err
	}

	// Generate key pair
	var privateKey, publicKey string
	if clientPrivateKey != "" {
		// Validate provided private key
		if err := wireguard.ValidatePrivateKey(clientPrivateKey); err != nil {
			return nil, err
		}
		privateKey = clientPrivateKey
		publicKey, err = wireguard.GeneratePublicKey(privateKey)
		if err != nil {
			return nil, err
		}
	} else {
		// Auto-generate key pair
		privateKey, publicKey, err = wireguard.GenerateKeyPair()
		if err != nil {
			return nil, err
		}
	}

	// Generate peer ID
	peerID, err := snowflake.GenerateID()
	if err != nil {
		return nil, errors.WithCode(code.ErrWGPeerIDGenerationFailed, "failed to generate peer ID")
	}

	// Get global config for calculating effective endpoint and DNS
	cfg := config.Get()
	var wgOpts *options.WireGuardOptions
	if cfg != nil && cfg.WireGuard != nil {
		wgOpts = cfg.WireGuard
	}

	// Create a temporary peer model for calculating effective values
	tempPeer := &model.WGPeer{
		DNS:      dns,
		Endpoint: endpoint,
		IPPoolID: ipPoolID,
	}

	// Calculate effective endpoint and DNS using helper functions
	// This ensures we store the complete values (including defaults) in the database
	var effectiveEndpoint, effectiveDNS string
	if wgOpts != nil {
		effectiveEndpoint = CalculateEffectiveEndpoint(tempPeer, pool, wgOpts, w.configManager, ctx)
		effectiveDNS = CalculateEffectiveDNS(tempPeer, pool, wgOpts)
	} else {
		// Fallback to provided values if config is not available
		effectiveEndpoint = endpoint
		effectiveDNS = dns
	}

	// Create peer model with calculated effective values
	peer := &model.WGPeer{
		ID:                  peerID,
		UserID:              userID,
		DeviceName:          deviceName,
		ClientPrivateKey:    privateKey,
		ClientPublicKey:     publicKey,
		ClientIP:            clientIPCIDR,
		AllowedIPs:          allowedIPs,
		DNS:                 effectiveDNS,
		Endpoint:            effectiveEndpoint,
		PersistentKeepalive: 25,
		Status:              model.WGPeerStatusActive,
		IPPoolID:            ipPoolID,
	}

	if persistentKeepalive != nil {
		peer.PersistentKeepalive = *persistentKeepalive
	}

	// Create IP allocation record
	allocationID, err := snowflake.GenerateID()
	if err != nil {
		return nil, errors.WithCode(code.ErrWGPeerIDGenerationFailed, "failed to generate allocation ID")
	}

	allocation := &model.IPAllocation{
		ID:        allocationID,
		IPPoolID:  ipPoolID,
		PeerID:    peerID,
		IPAddress: allocatedIP,
		Status:    model.IPAllocationStatusAllocated,
	}

	// Save to database
	if err := w.store.WGPeers().CreatePeer(ctx, peer); err != nil {
		return nil, err
	}

	if err := w.store.IPAllocations().CreateIPAllocation(ctx, allocation); err != nil {
		// Rollback: delete peer if allocation fails
		_ = w.store.WGPeers().DeletePeer(ctx, peerID)
		return nil, err
	}

	// Generate and save client config file
	if err := w.generateAndSaveClientConfig(ctx, peer); err != nil {
		klog.V(1).InfoS("failed to generate client config", "peerID", peerID, "error", err)
		// Continue anyway, config file generation is not critical
	}

	// Update server config file
	if w.configManager != nil {
		if err := w.updateServerConfigForPeer(peer, true); err != nil {
			klog.V(1).InfoS("failed to update server config", "peerID", peerID, "error", err)
			// Continue anyway, server config update failure is logged but doesn't block
		} else {
			// Apply server config
			if err := w.configManager.ApplyConfig(); err != nil {
				klog.V(1).InfoS("failed to apply server config", "peerID", peerID, "error", err)
				// Continue anyway
			}
		}
	}

	return peer, nil
}

func (w *wgPeerSrv) GetPeer(ctx context.Context, id string) (*model.WGPeer, error) {
	return w.store.WGPeers().GetPeer(ctx, id)
}

func (w *wgPeerSrv) GetPeerByPublicKey(ctx context.Context, publicKey string) (*model.WGPeer, error) {
	return w.store.WGPeers().GetPeerByPublicKey(ctx, publicKey)
}

func (w *wgPeerSrv) UpdatePeer(ctx context.Context, peer *model.WGPeer, newClientIP, newIPPoolID *string) error {
	// Get existing peer to check what changed
	existingPeer, err := w.store.WGPeers().GetPeer(ctx, peer.ID)
	if err != nil {
		return err
	}

	// Handle IP address change
	if newClientIP != nil && *newClientIP != "" {
		// Extract IP from existing CIDR format
		existingIP, _ := ip.ExtractIPFromCIDR(existingPeer.ClientIP)
		
		// Check if IP actually changed
		if existingIP != *newClientIP {
			// Determine which IP pool to use
			ipPoolID := peer.IPPoolID
			if newIPPoolID != nil && *newIPPoolID != "" {
				ipPoolID = *newIPPoolID
			} else if ipPoolID == "" {
				// Get default IP pool if not specified
				pools, _, err := w.store.IPPools().ListIPPools(ctx, store.IPPoolListOptions{
					Status: model.IPPoolStatusActive,
					Limit:  1,
				})
				if err != nil || len(pools) == 0 {
					return errors.WithCode(code.ErrIPPoolNotFound, "no active IP pool found")
				}
				ipPoolID = pools[0].ID
				peer.IPPoolID = ipPoolID
			}

			// Get server tunnel IP from server config Address
			serverTunnelIP := ""
			if w.configManager != nil {
				serverConfig, err := w.configManager.ReadServerConfig()
				if err == nil && serverConfig != nil && serverConfig.Interface != nil && serverConfig.Interface.Address != "" {
					// Extract IP from Address (e.g., "100.100.100.1/24" -> "100.100.100.1")
					extractedIP, err := ip.ExtractIPFromCIDR(serverConfig.Interface.Address)
					if err == nil {
						serverTunnelIP = extractedIP
					}
				}
			}

			// Validate and allocate new IP
			allocator := ip.NewAllocator(w.store)
			if err := allocator.ValidateAndAllocateIP(ctx, ipPoolID, *newClientIP, serverTunnelIP); err != nil {
				return err
			}

			// Release old IP allocation
			if err := allocator.ReleaseIP(ctx, peer.ID); err != nil {
				klog.V(1).InfoS("failed to release old IP allocation", "peerID", peer.ID, "error", err)
				// Continue anyway
			}

			// Create new IP allocation record
			allocationID, err := snowflake.GenerateID()
			if err != nil {
				return errors.WithCode(code.ErrWGPeerIDGenerationFailed, "failed to generate allocation ID")
			}

			allocation := &model.IPAllocation{
				ID:        allocationID,
				IPPoolID:  ipPoolID,
				PeerID:    peer.ID,
				IPAddress: *newClientIP,
				Status:    model.IPAllocationStatusAllocated,
			}

			if err := w.store.IPAllocations().CreateIPAllocation(ctx, allocation); err != nil {
				return err
			}

			// Format IP as CIDR
			clientIPCIDR, err := ip.FormatIPAsCIDR(*newClientIP)
			if err != nil {
				return err
			}
			peer.ClientIP = clientIPCIDR
		}
	}

	// Handle IP Pool change (without IP change)
	// Allow clearing IP Pool by setting it to empty string
	ipPoolChanged := false
	if newIPPoolID != nil && peer.IPPoolID != *newIPPoolID {
		peer.IPPoolID = *newIPPoolID
		ipPoolChanged = true
	}

	// Handle private key change - regenerate public key if needed
	if peer.ClientPrivateKey != existingPeer.ClientPrivateKey && peer.ClientPrivateKey != "" {
		// Validate private key
		if err := wireguard.ValidatePrivateKey(peer.ClientPrivateKey); err != nil {
			return errors.WithCode(code.ErrWGPrivateKeyInvalid, "invalid private key: %s", err.Error())
		}
		// Public key should already be updated in controller, but regenerate to be safe
		publicKey, err := wireguard.GeneratePublicKey(peer.ClientPrivateKey)
		if err != nil {
			return errors.WithCode(code.ErrWGPublicKeyGenerationFailed, "failed to generate public key: %s", err.Error())
		}
		peer.ClientPublicKey = publicKey
	}

	// Recalculate effective endpoint and DNS if needed:
	// 1. Endpoint or DNS is empty (needs default value)
	// 2. IP Pool changed (may have different pool config)
	needsRecalculation := peer.Endpoint == "" || peer.DNS == "" || ipPoolChanged

	if needsRecalculation {
		// Get global config for calculating effective endpoint and DNS
		cfg := config.Get()
		var wgOpts *options.WireGuardOptions
		if cfg != nil && cfg.WireGuard != nil {
			wgOpts = cfg.WireGuard
		}

		// Get IP pool if peer has IPPoolID
		var pool *model.IPPool
		if peer.IPPoolID != "" {
			var err error
			pool, err = w.store.IPPools().GetIPPool(ctx, peer.IPPoolID)
			if err != nil {
				klog.V(1).InfoS("failed to get IP pool for recalculation", "poolID", peer.IPPoolID, "error", err)
				// Continue without pool config
			}
		}

		// Calculate effective endpoint and DNS using helper functions
		if wgOpts != nil {
			// Only recalculate if the field is empty or IP Pool changed
			if peer.Endpoint == "" || ipPoolChanged {
				peer.Endpoint = CalculateEffectiveEndpoint(peer, pool, wgOpts, w.configManager, ctx)
			}
			if peer.DNS == "" || ipPoolChanged {
				peer.DNS = CalculateEffectiveDNS(peer, pool, wgOpts)
			}
		}
	}

	// Update database
	if err := w.store.WGPeers().UpdatePeer(ctx, peer); err != nil {
		return err
	}

	// Update server config if status or AllowedIPs changed
	if w.configManager != nil {
		statusChanged := existingPeer.Status != peer.Status
		allowedIPsChanged := existingPeer.AllowedIPs != peer.AllowedIPs
		persistentKeepaliveChanged := existingPeer.PersistentKeepalive != peer.PersistentKeepalive

		if statusChanged || allowedIPsChanged || persistentKeepaliveChanged {
			// Only update if peer is active
			if peer.Status == model.WGPeerStatusActive {
				if err := w.updateServerConfigForPeer(peer, false); err != nil {
					klog.V(1).InfoS("failed to update server config", "peerID", peer.ID, "error", err)
					// Continue anyway
				} else {
					// Apply server config
					if err := w.configManager.ApplyConfig(); err != nil {
						klog.V(1).InfoS("failed to apply server config", "peerID", peer.ID, "error", err)
						// Continue anyway
					}
				}
			} else {
				// Peer is disabled, remove from server config
				if err := w.configManager.RemovePeer(peer.ClientPublicKey); err != nil {
					klog.V(1).InfoS("failed to remove peer from server config", "peerID", peer.ID, "error", err)
					// Continue anyway
				} else {
					// Apply server config
					if err := w.configManager.ApplyConfig(); err != nil {
						klog.V(1).InfoS("failed to apply server config", "peerID", peer.ID, "error", err)
						// Continue anyway
					}
				}
			}
		}
	}

	// Regenerate client config if needed
	if err := w.generateAndSaveClientConfig(ctx, peer); err != nil {
		klog.V(1).InfoS("failed to regenerate client config", "peerID", peer.ID, "error", err)
		// Continue anyway
	}

	return nil
}

func (w *wgPeerSrv) DeletePeer(ctx context.Context, id string, isHardDelete bool) error {
	// Get peer before deletion to remove from server config
	peer, err := w.store.WGPeers().GetPeer(ctx, id)
	if err != nil {
		// If peer not found, still try to release/delete IP and continue
		klog.V(1).InfoS("peer not found, continuing with deletion", "peerID", id, "error", err)
	} else if w.configManager != nil {
		// Remove peer from server config
		if err := w.configManager.RemovePeer(peer.ClientPublicKey); err != nil {
			klog.V(1).InfoS("failed to remove peer from server config", "peerID", id, "error", err)
			// Continue anyway
		} else {
			// Apply server config
			if err := w.configManager.ApplyConfig(); err != nil {
				klog.V(1).InfoS("failed to apply server config", "peerID", id, "error", err)
				// Continue anyway
			}
		}
	}

	// Delete client config file
	cfg := config.Get()
	if cfg != nil && cfg.WireGuard != nil {
		userDir := cfg.WireGuard.ResolvedUserDir()
		configPath := filepath.Join(userDir, id+".conf")
		if err := os.Remove(configPath); err != nil {
			if !os.IsNotExist(err) {
				// Only log if file exists but deletion failed
				klog.V(1).InfoS("failed to delete client config file", "peerID", id, "path", configPath, "error", err)
			}
			// Continue anyway, file deletion is not critical
		} else {
			klog.V(2).InfoS("deleted client config file", "peerID", id, "path", configPath)
		}
	}

	// Handle IP allocation: hard delete for admin, soft delete for regular users
	if isHardDelete {
		// Hard delete: remove IP allocation record completely
		if err := w.store.IPAllocations().DeleteIPAllocationByPeerID(ctx, id); err != nil {
			// Log error but continue with deletion
			klog.V(1).InfoS("failed to hard delete IP allocation", "peerID", id, "error", err)
		}
	} else {
		// Soft delete: mark IP allocation as released
		if err := w.ReleaseIP(ctx, id); err != nil {
			// Log error but continue with deletion
			klog.V(1).InfoS("failed to release IP allocation", "peerID", id, "error", err)
		}
	}

	return w.store.WGPeers().DeletePeer(ctx, id)
}

// ReleaseIP releases the IP allocation for a peer.
func (w *wgPeerSrv) ReleaseIP(ctx context.Context, peerID string) error {
	allocator := ip.NewAllocator(w.store)
	return allocator.ReleaseIP(ctx, peerID)
}

func (w *wgPeerSrv) ListPeers(ctx context.Context, opt store.WGPeerListOptions) ([]*model.WGPeer, int64, error) {
	return w.store.WGPeers().ListPeers(ctx, opt)
}

func (w *wgPeerSrv) CountPeersByUserID(ctx context.Context, userID string) (int64, error) {
	return w.store.WGPeers().CountPeersByUserID(ctx, userID)
}

// UpdatePeersForIPPoolChange updates all peers that use a specific IP pool
// when the pool's Endpoint or DNS changes.
func (w *wgPeerSrv) UpdatePeersForIPPoolChange(ctx context.Context, poolID string, newPool *model.IPPool) error {
	// Find all peers using this pool
	peers, _, err := w.store.WGPeers().ListPeers(ctx, store.WGPeerListOptions{
		IPPoolID: poolID,
	})
	if err != nil {
		return err
	}

	// Get global config for calculating effective endpoint and DNS
	cfg := config.Get()
	var wgOpts *options.WireGuardOptions
	if cfg != nil && cfg.WireGuard != nil {
		wgOpts = cfg.WireGuard
	}

	// Update each peer if needed
	for _, peer := range peers {
		needsUpdate := false

		// If peer's Endpoint is empty, it uses pool's default, so recalculate
		if peer.Endpoint == "" {
			if wgOpts != nil {
				peer.Endpoint = CalculateEffectiveEndpoint(peer, newPool, wgOpts, w.configManager, ctx)
				needsUpdate = true
			}
		}

		// If peer's DNS is empty, it uses pool's default, so recalculate
		if peer.DNS == "" {
			if wgOpts != nil {
				peer.DNS = CalculateEffectiveDNS(peer, newPool, wgOpts)
				needsUpdate = true
			}
		}

		if needsUpdate {
			// Update peer in database
			if err := w.store.WGPeers().UpdatePeer(ctx, peer); err != nil {
				klog.V(1).InfoS("failed to update peer after pool change", "peerID", peer.ID, "error", err)
				// Continue with other peers
				continue
			}

			// Regenerate client config file
			if err := w.generateAndSaveClientConfig(ctx, peer); err != nil {
				klog.V(1).InfoS("failed to regenerate client config after pool change", "peerID", peer.ID, "error", err)
				// Continue with other peers
			}
		}
	}

	return nil
}

// generateAndSaveClientConfig generates and saves the client configuration file.
func (w *wgPeerSrv) generateAndSaveClientConfig(ctx context.Context, peer *model.WGPeer) error {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return errors.WithCode(code.ErrWGConfigNotInitialized, "WireGuard config not initialized")
	}

	wgOpts := cfg.WireGuard

	// Get server public key
	var serverPublicKey string
	if w.configManager != nil {
		var err error
		serverPublicKey, err = w.configManager.GetServerPublicKey()
		if err != nil {
			return errors.Wrap(err, "failed to get server public key")
		}
	}

	// Get server MTU from server config
	var mtu int
	if w.configManager != nil {
		serverConfig, err := w.configManager.ReadServerConfig()
		if err == nil && serverConfig != nil && serverConfig.Interface != nil {
			if serverConfig.Interface.MTU > 0 {
				mtu = serverConfig.Interface.MTU
			}
		}
	}

	// Use values directly from database - they are already calculated and stored during create/update
	// DNS and Endpoint are guaranteed to have default values (if applicable) stored in the database
	dns := peer.DNS
	endpoint := peer.Endpoint
	allowedIPs := peer.AllowedIPs

	// Only calculate AllowedIPs default if it's still empty (shouldn't happen for new peers, but handle legacy data)
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

	klog.V(2).InfoS("client config file saved", "peerID", peer.ID, "path", configPath)
	return nil
}

// updateServerConfigForPeer updates the server configuration for a peer.
func (w *wgPeerSrv) updateServerConfigForPeer(peer *model.WGPeer, isNew bool) error {
	if w.configManager == nil {
		return errors.WithCode(code.ErrWGConfigNotInitialized, "config manager not initialized")
	}

	// Only update if peer is active
	if peer.Status != model.WGPeerStatusActive {
		return nil
	}

	serverPeer := &wireguard.ServerPeerConfig{
		PublicKey:           peer.ClientPublicKey,
		AllowedIPs:          peer.ClientIP, // Use ClientIP as AllowedIPs in server config
		PersistentKeepalive: peer.PersistentKeepalive,
		Comment:             peer.DeviceName,
	}

	if isNew {
		return w.configManager.AddPeer(serverPeer)
	}
	return w.configManager.UpdatePeer(peer.ClientPublicKey, serverPeer)
}

// CalculateEffectiveEndpoint calculates the effective endpoint for a peer.
// Priority: peer.Endpoint > pool.Endpoint > ServerIP:ListenPort > wgOpts.Endpoint
func CalculateEffectiveEndpoint(peer *model.WGPeer, pool *model.IPPool, wgOpts *options.WireGuardOptions, configManager *wireguard.ServerConfigManager, ctx context.Context) string {
	endpoint := peer.Endpoint
	if endpoint == "" && pool != nil && pool.Endpoint != "" {
		endpoint = pool.Endpoint
	}
	if endpoint == "" {
		// Use Settings.ServerIP + ListenPort from server config
		if configManager != nil {
			serverConfig, err := configManager.ReadServerConfig()
			if err == nil && serverConfig != nil && serverConfig.Interface != nil {
				serverIP := wgOpts.ServerIP
				if serverIP == "" {
					// Auto-detect if not set
					detectedIP, err := network.GetServerIP(ctx, "")
					if err == nil {
						serverIP = detectedIP
					}
				}
				if serverIP != "" && serverConfig.Interface.ListenPort > 0 {
					endpoint = fmt.Sprintf("%s:%d", serverIP, serverConfig.Interface.ListenPort)
				}
			}
		}
		// Fallback to global endpoint if still empty
		if endpoint == "" {
			endpoint = wgOpts.Endpoint
		}
	}
	return endpoint
}

// CalculateEffectiveDNS calculates the effective DNS for a peer.
// Priority: peer.DNS > pool.DNS (if pool exists, even if empty) > wgOpts.DNS (if pool doesn't exist)
func CalculateEffectiveDNS(peer *model.WGPeer, pool *model.IPPool, wgOpts *options.WireGuardOptions) string {
	dns := peer.DNS
	if dns == "" && pool != nil {
		// If associated with IP Pool, only use IP Pool DNS (even if empty, don't fallback)
		dns = pool.DNS
	} else if dns == "" && pool == nil {
		// Only when Peer is not associated with IP Pool, use Settings/Global config DNS
		dns = wgOpts.DNS
	}
	return dns
}

// CalculateIPPoolEndpoint calculates the default endpoint for an IP pool.
// Priority: poolEndpoint > ServerIP:ListenPort > wgOpts.Endpoint
func CalculateIPPoolEndpoint(poolEndpoint string, wgOpts *options.WireGuardOptions, configManager *wireguard.ServerConfigManager, ctx context.Context) string {
	endpoint := poolEndpoint
	if endpoint == "" {
		// Use Settings.ServerIP + ListenPort from server config
		if configManager != nil {
			serverConfig, err := configManager.ReadServerConfig()
			if err == nil && serverConfig != nil && serverConfig.Interface != nil {
				serverIP := wgOpts.ServerIP
				if serverIP == "" {
					// Auto-detect if not set
					detectedIP, err := network.GetServerIP(ctx, "")
					if err == nil {
						serverIP = detectedIP
					}
				}
				if serverIP != "" && serverConfig.Interface.ListenPort > 0 {
					endpoint = fmt.Sprintf("%s:%d", serverIP, serverConfig.Interface.ListenPort)
				}
			}
		}
		// Fallback to global endpoint if still empty
		if endpoint == "" {
			endpoint = wgOpts.Endpoint
		}
	}
	return endpoint
}

// UpdatePeersEndpointForGlobalConfigChange updates all peers that use default endpoint
// when global config (ServerIP, ListenPort, or Endpoint) changes.
func (w *wgPeerSrv) UpdatePeersEndpointForGlobalConfigChange(ctx context.Context) error {
	// Get all peers
	peers, _, err := w.store.WGPeers().ListPeers(ctx, store.WGPeerListOptions{
		Limit: 10000, // Get all peers
	})
	if err != nil {
		return err
	}

	// Get global config
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return nil
	}
	wgOpts := cfg.WireGuard

	// Get server IP
	serverIP := wgOpts.ServerIP
	if serverIP == "" {
		detectedIP, err := network.GetServerIP(ctx, "")
		if err == nil {
			serverIP = detectedIP
		}
	}

	// Get current ListenPort from server config
	var currentListenPort int
	if w.configManager != nil {
		serverConfig, err := w.configManager.ReadServerConfig()
		if err == nil && serverConfig != nil && serverConfig.Interface != nil {
			currentListenPort = serverConfig.Interface.ListenPort
		}
	}

	// Update each peer if needed
	for _, peer := range peers {
		needsUpdate := false

		// Check if peer uses default endpoint
		if peer.Endpoint == "" {
			// Empty endpoint means using default
			needsUpdate = true
		} else {
			// Check if endpoint matches server IP pattern (ServerIP:Port)
			endpointIP, err := ip.ExtractIPFromEndpoint(peer.Endpoint)
			if err == nil && endpointIP != "" {
				// Check if it matches current server IP
				if serverIP != "" && endpointIP == serverIP {
					// Endpoint uses server IP, check if port matches current ListenPort
					if currentListenPort > 0 {
						// Extract port from current endpoint
						_, port, err := net.SplitHostPort(peer.Endpoint)
						if err == nil {
							currentPort, err := strconv.Atoi(port)
							if err == nil {
								if currentPort != currentListenPort {
									// Port doesn't match, needs update
									needsUpdate = true
								}
							} else {
								// Can't parse port, assume needs update
								needsUpdate = true
							}
						} else {
							// Can't parse endpoint, assume needs update
							needsUpdate = true
						}
					} else {
						// Can't get ListenPort, assume needs update
						needsUpdate = true
					}
				} else if peer.Endpoint == wgOpts.Endpoint {
					// Endpoint equals global endpoint config
					needsUpdate = true
				}
			}
		}

		if needsUpdate {
			// Get IP pool if peer has one
			var pool *model.IPPool
			if peer.IPPoolID != "" {
				pool, _ = w.store.IPPools().GetIPPool(ctx, peer.IPPoolID)
			}

			// Recalculate endpoint directly (don't rely on CalculateEffectiveEndpoint
			// which returns early if peer.Endpoint is not empty)
			var newEndpoint string
			if peer.Endpoint == "" {
				// Use CalculateEffectiveEndpoint for empty endpoint
				newEndpoint = CalculateEffectiveEndpoint(peer, pool, wgOpts, w.configManager, ctx)
			} else {
				// For non-empty endpoint, recalculate based on current ServerIP and ListenPort
				endpointIP, _ := ip.ExtractIPFromEndpoint(peer.Endpoint)
				if endpointIP != "" && serverIP != "" && endpointIP == serverIP {
					// Use current ServerIP with current ListenPort
					if currentListenPort > 0 {
						newEndpoint = fmt.Sprintf("%s:%d", serverIP, currentListenPort)
					} else {
						// Fallback to global endpoint if can't get ListenPort
						newEndpoint = wgOpts.Endpoint
					}
				} else if peer.Endpoint == wgOpts.Endpoint {
					// Use global endpoint (which may have changed)
					newEndpoint = wgOpts.Endpoint
				} else {
					// Endpoint doesn't match server IP or global config, don't update
					continue
				}
			}

			// Only update if endpoint actually changed
			if newEndpoint != "" && newEndpoint != peer.Endpoint {
				peer.Endpoint = newEndpoint
				// Update in database
				if err := w.store.WGPeers().UpdatePeer(ctx, peer); err != nil {
					klog.V(1).InfoS("failed to update peer endpoint", "peerID", peer.ID, "error", err)
					continue
				}

				// Regenerate client config
				if err := w.generateAndSaveClientConfig(ctx, peer); err != nil {
					klog.V(1).InfoS("failed to regenerate client config", "peerID", peer.ID, "error", err)
				}
			}
		}
	}

	return nil
}

// UpdatePeersDNSForGlobalConfigChange updates all peers that use default DNS
// when global config DNS changes.
func (w *wgPeerSrv) UpdatePeersDNSForGlobalConfigChange(ctx context.Context) error {
	// Get all peers
	peers, _, err := w.store.WGPeers().ListPeers(ctx, store.WGPeerListOptions{
		Limit: 10000, // Get all peers
	})
	if err != nil {
		return err
	}

	// Get global config
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return nil
	}
	wgOpts := cfg.WireGuard

	// Update each peer if needed
	for _, peer := range peers {
		// Check if peer uses default DNS (empty means using default)
		if peer.DNS == "" {
			// Get IP pool if peer has one
			var pool *model.IPPool
			if peer.IPPoolID != "" {
				pool, _ = w.store.IPPools().GetIPPool(ctx, peer.IPPoolID)
			}

			// Recalculate DNS
			oldDNS := peer.DNS
			peer.DNS = CalculateEffectiveDNS(peer, pool, wgOpts)

			// Only update if DNS actually changed
			if oldDNS != peer.DNS {
				// Update in database
				if err := w.store.WGPeers().UpdatePeer(ctx, peer); err != nil {
					klog.V(1).InfoS("failed to update peer DNS", "peerID", peer.ID, "error", err)
					continue
				}

				// Regenerate client config
				if err := w.generateAndSaveClientConfig(ctx, peer); err != nil {
					klog.V(1).InfoS("failed to regenerate client config", "peerID", peer.ID, "error", err)
				}
			}
		}
	}

	return nil
}
