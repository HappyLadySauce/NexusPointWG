package service

import (
	"context"
	"net/netip"
	"strings"

	"github.com/HappyLadySauce/NexusPointWG/internal/local"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	ipalloc "github.com/HappyLadySauce/NexusPointWG/internal/pkg/ipalloc"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	wgfile "github.com/HappyLadySauce/NexusPointWG/internal/pkg/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/snowflake"
	"github.com/HappyLadySauce/errors"
)

// ListPeers lists all WireGuard peers with filtering and pagination.
func (w *wgSrv) ListPeers(ctx context.Context, opt store.WGPeerListOptions) ([]*model.WGPeer, int64, error) {
	return w.store.WGPeers().List(ctx, opt)
}

// GetPeer returns a peer by ID.
func (w *wgSrv) GetPeer(ctx context.Context, id string) (*model.WGPeer, error) {
	return w.store.WGPeers().Get(ctx, id)
}

// CreatePeer creates a new WireGuard peer for a user.
func (w *wgSrv) CreatePeer(ctx context.Context, params CreatePeerParams) (*model.WGPeer, error) {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return nil, errors.WithCode(code.ErrWGConfigNotInitialized, "")
	}

	// Determine endpoint: use request endpoint if provided, otherwise use config default
	endpoint := cfg.WireGuard.Endpoint
	if params.Endpoint != nil && strings.TrimSpace(*params.Endpoint) != "" {
		endpoint = strings.TrimSpace(*params.Endpoint)
	}
	if strings.TrimSpace(endpoint) == "" {
		return nil, errors.WithCode(code.ErrWGEndpointRequired, "")
	}

	// TODO: File locking will be handled later

	owner, err := w.store.Users().GetUserByUsername(ctx, params.Username)
	if err != nil {
		return nil, err
	}

	// Get server configuration info using structured parsing
	serverPub, serverIPStr, mtu, allocationPrefix, err := w.extractServerInfo(ctx)
	if err != nil {
		return nil, err
	}

	serverIP, err := netip.ParseAddr(serverIPStr)
	if err != nil {
		return nil, errors.WithCode(code.ErrWGServerAddressInvalid, "invalid server IP: %v", err)
	}

	// Validate prefix: must have enough bits for allocation
	if allocationPrefix.Bits() >= 30 {
		maxHosts := 1 << (32 - allocationPrefix.Bits())
		usableHosts := maxHosts - 2 // subtract network and broadcast
		if allocationPrefix.Bits() == 32 {
			usableHosts = 0 // /32 has no usable hosts for allocation
		}
		return nil, errors.WithCode(code.ErrWGPrefixTooSmall,
			"AllowedIPs prefix too small (%s): only %d usable host(s), need at least /29 (e.g., 100.100.100.0/24) for client IP allocation", allocationPrefix.String(), usableHosts)
	}

	// Collect used IPs from database
	peers, _, err := w.ListPeers(ctx, store.WGPeerListOptions{Limit: 10000})
	if err != nil {
		return nil, err
	}
	usedIPs := ipalloc.CollectUsedIPsFromPeers(peers, allocationPrefix)

	// Create allocator and allocate IP
	allocator := ipalloc.NewAllocator(allocationPrefix, serverIP, usedIPs)
	clientAddr, err := allocator.Allocate()
	if err != nil {
		return nil, errors.WithCode(code.ErrWGIPAllocationFailed, "failed to allocate IP: %v", err)
	}
	clientCIDR := clientAddr.String() + "/32"

	// Use provided private key or generate one
	var clientPriv string
	if params.PrivateKey != nil && strings.TrimSpace(*params.PrivateKey) != "" {
		clientPriv = strings.TrimSpace(*params.PrivateKey)
		// Validate the private key
		if err := wgfile.ValidatePrivateKey(ctx, clientPriv); err != nil {
			return nil, errors.WithCode(code.ErrWGPrivateKeyInvalid, "invalid private key: %v", err)
		}
	} else {
		// Auto-generate private key
		clientPriv, err = wgfile.GeneratePrivateKey(ctx)
		if err != nil {
			return nil, err
		}
	}

	clientPub, err := wgfile.DerivePublicKey(ctx, clientPriv)
	if err != nil {
		return nil, err
	}

	peerID, err := snowflake.GenerateID()
	if err != nil {
		return nil, errors.WithCode(code.ErrWGPeerIDGenerationFailed, "failed to generate id: %v", err)
	}

	keepalive := 0
	if params.PersistentKeepalive != nil {
		keepalive = *params.PersistentKeepalive
	}

	dns := ""
	if params.DNS != nil {
		dns = strings.TrimSpace(*params.DNS)
	}
	peer := &model.WGPeer{
		ID:                  peerID,
		UserID:              owner.ID,
		DeviceName:          params.DeviceName,
		ClientPublicKey:     clientPub,
		ClientIP:            clientCIDR,
		AllowedIPs:          strings.TrimSpace(params.AllowedIPs),
		DNS:                 dns,
		PersistentKeepalive: keepalive,
		Status:              model.WGPeerStatusActive,
	}

	// 1) Write user files first (derived artifacts)
	if err := w.writeUserFiles(ctx, owner.Username, peer, clientPriv, serverPub, mtu, endpoint); err != nil {
		return nil, err
	}

	// 2) Store peer (source of truth)
	if err := w.store.WGPeers().Create(ctx, peer); err != nil {
		// best-effort cleanup of derived artifacts
		userFilesStore := local.NewLocalStore().UserConfigStore()
		_ = userFilesStore.Delete(ctx, owner.Username, peer.ID)
		return nil, err
	}

	// 3) Apply server config
	if err := w.SyncServerConfig(ctx); err != nil {
		return nil, err
	}

	return peer, nil
}

// updatePeerFields updates peer fields from request, returns true if user files need regeneration.
func (w *wgSrv) updatePeerFields(peer *model.WGPeer, params UpdatePeerParams) (regenerateUserFiles bool, privateKeyChanged bool, newPrivateKey string, endpoint string, err error) {
	cfg := config.Get()
	endpoint = cfg.WireGuard.Endpoint

	// Admin update - can update all fields
	if params.AllowedIPs != nil {
		newAllowedIPs := strings.TrimSpace(*params.AllowedIPs)
		if peer.AllowedIPs != newAllowedIPs {
			peer.AllowedIPs = newAllowedIPs
			regenerateUserFiles = true
		}
	}
	if params.PersistentKeepalive != nil {
		if peer.PersistentKeepalive != *params.PersistentKeepalive {
			peer.PersistentKeepalive = *params.PersistentKeepalive
			regenerateUserFiles = true
		}
	}
	if params.DNS != nil {
		newDNS := strings.TrimSpace(*params.DNS)
		if peer.DNS != newDNS {
			peer.DNS = newDNS
			regenerateUserFiles = true
		}
	}
	if params.Status != nil {
		peer.Status = *params.Status
	}
	if params.DeviceName != nil {
		peer.DeviceName = strings.TrimSpace(*params.DeviceName)
	}
	if params.PrivateKey != nil {
		newPriv := strings.TrimSpace(*params.PrivateKey)
		if err := wgfile.ValidatePrivateKey(context.Background(), newPriv); err != nil {
			return false, false, "", "", errors.WithCode(code.ErrWGPrivateKeyInvalid, "invalid private key: %v", err)
		}
		privateKeyChanged = true
		newPrivateKey = newPriv
		regenerateUserFiles = true
	}
	if params.ClientIP != nil {
		newClientIP := strings.TrimSpace(*params.ClientIP)
		if peer.ClientIP != newClientIP {
			peer.ClientIP = newClientIP
			regenerateUserFiles = true
		}
	}
	if params.Endpoint != nil && strings.TrimSpace(*params.Endpoint) != "" {
		endpoint = strings.TrimSpace(*params.Endpoint)
	}

	return regenerateUserFiles, privateKeyChanged, newPrivateKey, endpoint, nil
}

// regenerateUserFilesIfNeeded regenerates user files if needed.
func (w *wgSrv) regenerateUserFilesIfNeeded(ctx context.Context, peer *model.WGPeer, regenerateUserFiles bool, privateKeyChanged bool, newPrivateKey string, endpoint string) error {
	if !regenerateUserFiles {
		return nil
	}

	user, err := w.store.Users().GetUser(ctx, peer.UserID)
	if err != nil {
		return err
	}

	serverPub, _, mtu, _, err := w.extractServerInfo(ctx)
	if err != nil {
		return err
	}

	// Read or use new client private key
	var clientPriv string
	if privateKeyChanged {
		clientPriv = newPrivateKey
		// Update public key
		newPub, err := wgfile.DerivePublicKey(ctx, newPrivateKey)
		if err != nil {
			return err
		}
		peer.ClientPublicKey = newPub
	} else {
		clientFiles, err := w.readUserFiles(ctx, user.Username, peer.ID)
		if err != nil {
			// If we can't read the files, we can't regenerate them
			// This is a critical error for updates that require file regeneration
			return errors.WithCode(code.ErrWGUserConfigNotFound, "failed to read user files for regeneration: %v", err)
		}
		if clientFiles == nil || clientFiles.PrivateKey == "" {
			return errors.WithCode(code.ErrWGPrivateKeyReadFailed, "private key not found in user files")
		}
		clientPriv = clientFiles.PrivateKey
	}

	if clientPriv == "" {
		return errors.WithCode(code.ErrWGPrivateKeyReadFailed, "client private key is empty")
	}

	return w.writeUserFiles(ctx, user.Username, peer, clientPriv, serverPub, mtu, endpoint)
}

// UpdatePeer updates a peer.
func (w *wgSrv) UpdatePeer(ctx context.Context, id string, params UpdatePeerParams) (*model.WGPeer, error) {
	peer, err := w.store.WGPeers().Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate ClientIP if changed
	if params.ClientIP != nil {
		newClientIP := strings.TrimSpace(*params.ClientIP)
		// Parse the new ClientIP
		prefix, err := netip.ParsePrefix(newClientIP)
		if err != nil {
			return nil, errors.WithCode(code.ErrValidation, "invalid ClientIP format: %v", err)
		}
		ip := prefix.Addr()
		if !ip.Is4() {
			return nil, errors.WithCode(code.ErrIPNotIPv4, code.Message(code.ErrIPNotIPv4))
		}

		// Get allocation prefix from server config
		_, serverIPStr, _, allocationPrefix, err := w.extractServerInfo(ctx)
		if err != nil {
			return nil, err
		}

		serverIP, err := netip.ParseAddr(serverIPStr)
		if err != nil {
			return nil, errors.WithCode(code.ErrWGServerAddressInvalid, "invalid server IP: %v", err)
		}

		// Collect used IPs (excluding current peer)
		peers, _, err := w.ListPeers(ctx, store.WGPeerListOptions{Limit: 10000})
		if err != nil {
			return nil, err
		}

		// Filter out current peer
		filteredPeers := make([]*model.WGPeer, 0, len(peers))
		for _, p := range peers {
			if p != nil && p.ID != id {
				filteredPeers = append(filteredPeers, p)
			}
		}

		usedIPs := ipalloc.CollectUsedIPsFromPeers(filteredPeers, allocationPrefix)

		// Create Allocator and validate
		allocator := ipalloc.NewAllocator(allocationPrefix, serverIP, usedIPs)
		if err := allocator.Validate(ip); err != nil {
			return nil, err
		}
	}

	// Update fields
	regenerateUserFiles, privateKeyChanged, newPrivateKey, endpoint, err := w.updatePeerFields(peer, params)
	if err != nil {
		return nil, err
	}

	// Save to database
	if err := w.store.WGPeers().Update(ctx, peer); err != nil {
		return nil, err
	}

	// Regenerate user files if needed
	if err := w.regenerateUserFilesIfNeeded(ctx, peer, regenerateUserFiles, privateKeyChanged, newPrivateKey, endpoint); err != nil {
		return nil, err
	}

	// Update public key if private key changed
	if privateKeyChanged {
		if err := w.store.WGPeers().Update(ctx, peer); err != nil {
			return nil, err
		}
	}

	// Apply changes to server config
	if err := w.SyncServerConfig(ctx); err != nil {
		return nil, err
	}
	return peer, nil
}

// DeletePeer deletes a peer.
func (w *wgSrv) DeletePeer(ctx context.Context, id string) error {
	peer, err := w.store.WGPeers().Get(ctx, id)
	if err != nil {
		return err
	}

	// Delete database record (hard delete)
	if err := w.store.WGPeers().Delete(ctx, id); err != nil {
		return err
	}

	// Delete configuration files (hard delete)
	user, err := w.store.Users().GetUser(ctx, peer.UserID)
	if err == nil {
		userFilesStore := local.NewLocalStore().UserConfigStore()
		_ = userFilesStore.Delete(ctx, user.Username, id)
	}

	// Sync server config to remove peer from server configuration
	if err := w.SyncServerConfig(ctx); err != nil {
		return err
	}

	return nil
}

// DownloadConfig downloads a peer configuration file.
func (w *wgSrv) DownloadConfig(ctx context.Context, peerID string) (string, []byte, error) {
	peer, err := w.store.WGPeers().Get(ctx, peerID)
	if err != nil {
		return "", nil, err
	}
	user, err := w.store.Users().GetUser(ctx, peer.UserID)
	if err != nil {
		return "", nil, err
	}

	userFilesStore := local.NewLocalStore().UserConfigStore()
	content, err := userFilesStore.Read(ctx, user.Username, peerID, "peer.conf")
	if err != nil {
		return "", nil, err
	}

	filename := sanitizeFilename(peer.DeviceName)
	if filename == "" {
		filename = peer.ID
	}
	return filename + ".conf", content, nil
}

// RotateConfig rotates keys for a peer.
func (w *wgSrv) RotateConfig(ctx context.Context, peerID string) error {
	peer, err := w.store.WGPeers().Get(ctx, peerID)
	if err != nil {
		return err
	}

	// TODO: File locking will be handled later

	// Generate new keys
	clientPriv, err := wgfile.GeneratePrivateKey(ctx)
	if err != nil {
		return err
	}
	clientPub, err := wgfile.DerivePublicKey(ctx, clientPriv)
	if err != nil {
		return err
	}

	user, err := w.store.Users().GetUser(ctx, peer.UserID)
	if err != nil {
		return err
	}

	// Update peer with new public key
	peer.ClientPublicKey = clientPub
	if err := w.store.WGPeers().Update(ctx, peer); err != nil {
		return err
	}

	// Extract server info and write user files
	serverPub, _, mtu, _, err := w.extractServerInfo(ctx)
	if err != nil {
		return err
	}

	cfg := config.Get()
	endpoint := cfg.WireGuard.Endpoint
	if err := w.writeUserFiles(ctx, user.Username, peer, clientPriv, serverPub, mtu, endpoint); err != nil {
		return err
	}

	// Sync server config
	return w.SyncServerConfig(ctx)
}

// sanitizeFilename sanitizes a filename for safe use.
func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "..", "_")
	return name
}
