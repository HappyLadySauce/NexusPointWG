package service

import (
	"context"
	"os"
	"path/filepath"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/core/ip"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/core/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/snowflake"
	"github.com/HappyLadySauce/errors"
	"k8s.io/klog/v2"
)

// WGPeerSrv defines the interface for WireGuard peer business logic.
type WGPeerSrv interface {
	CreatePeer(ctx context.Context, userID, deviceName, ipPoolID, clientIP, allowedIPs, dns, endpoint, clientPrivateKey string, persistentKeepalive *int) (*model.WGPeer, error)
	GetPeer(ctx context.Context, id string) (*model.WGPeer, error)
	GetPeerByPublicKey(ctx context.Context, publicKey string) (*model.WGPeer, error)
	UpdatePeer(ctx context.Context, peer *model.WGPeer) error
	DeletePeer(ctx context.Context, id string) error
	ListPeers(ctx context.Context, opt store.WGPeerListOptions) ([]*model.WGPeer, int64, error)
	ReleaseIP(ctx context.Context, peerID string) error
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
	if ipPoolID == "" {
		pools, _, err := w.store.IPPools().ListIPPools(ctx, store.IPPoolListOptions{
			Status: model.IPPoolStatusActive,
			Limit:  1,
		})
		if err != nil || len(pools) == 0 {
			return nil, errors.WithCode(code.ErrIPPoolNotFound, "no active IP pool found")
		}
		ipPoolID = pools[0].ID
	}

	// Allocate IP address
	allocator := ip.NewAllocator(w.store)
	var allocatedIP string
	if clientIP != "" {
		// Validate and use provided IP
		if err := allocator.ValidateAndAllocateIP(ctx, ipPoolID, clientIP); err != nil {
			return nil, err
		}
		allocatedIP = clientIP
	} else {
		// Auto-allocate IP
		var err error
		allocatedIP, err = allocator.AllocateIP(ctx, ipPoolID, "")
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

	// Create peer model
	peer := &model.WGPeer{
		ID:                  peerID,
		UserID:              userID,
		DeviceName:          deviceName,
		ClientPrivateKey:    privateKey,
		ClientPublicKey:     publicKey,
		ClientIP:            clientIPCIDR,
		AllowedIPs:          allowedIPs,
		DNS:                 dns,
		Endpoint:            endpoint,
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

func (w *wgPeerSrv) UpdatePeer(ctx context.Context, peer *model.WGPeer) error {
	// Get existing peer to check what changed
	existingPeer, err := w.store.WGPeers().GetPeer(ctx, peer.ID)
	if err != nil {
		return err
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

func (w *wgPeerSrv) DeletePeer(ctx context.Context, id string) error {
	// Get peer before deletion to remove from server config
	peer, err := w.store.WGPeers().GetPeer(ctx, id)
	if err != nil {
		// If peer not found, still try to release IP and continue
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

	// Release IP allocation before deleting peer
	if err := w.ReleaseIP(ctx, id); err != nil {
		// Log error but continue with deletion
		klog.V(1).InfoS("failed to release IP allocation", "peerID", id, "error", err)
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

	// Use defaults if peer fields are empty
	dns := peer.DNS
	if dns == "" {
		dns = wgOpts.DNS
	}

	endpoint := peer.Endpoint
	if endpoint == "" {
		endpoint = wgOpts.Endpoint
	}

	allowedIPs := peer.AllowedIPs
	if allowedIPs == "" {
		allowedIPs = wgOpts.DefaultAllowedIPs
	}

	// Generate client config
	clientConfig := &wireguard.ClientConfig{
		PrivateKey:          peer.ClientPrivateKey,
		Address:             peer.ClientIP,
		DNS:                 dns,
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
