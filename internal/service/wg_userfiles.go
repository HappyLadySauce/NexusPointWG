package service

import (
	"bytes"
	"context"
	"fmt"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	ipalloc "github.com/HappyLadySauce/NexusPointWG/internal/pkg/ipalloc"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	wgfile "github.com/HappyLadySauce/NexusPointWG/internal/pkg/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	iputil "github.com/HappyLadySauce/NexusPointWG/pkg/utils/ip"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/snowflake"
	"github.com/HappyLadySauce/errors"
)

func (w *wgSrv) AdminCreatePeer(ctx context.Context, req v1.CreateWGPeerRequest) (*model.WGPeer, error) {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return nil, errors.WithCode(code.ErrWGConfigNotInitialized, "")
	}

	// Determine endpoint: use request endpoint if provided, otherwise use config default
	endpoint := cfg.WireGuard.Endpoint
	if req.Endpoint != nil && strings.TrimSpace(*req.Endpoint) != "" {
		endpoint = strings.TrimSpace(*req.Endpoint)
	}
	if strings.TrimSpace(endpoint) == "" {
		return nil, errors.WithCode(code.ErrWGEndpointRequired, "")
	}

	// Serialize allocate-ip + file write + server apply.
	lockPath := filepath.Join(cfg.WireGuard.RootDir, ".nexuspointwg.lock")
	lock, err := wgfile.AcquireFileLock(lockPath)
	if err != nil {
		return nil, errors.WithCode(code.ErrWGLockAcquireFailed, "failed to acquire wireguard lock: %v", err)
	}
	defer func() { _ = lock.Release() }()

	owner, err := w.store.Users().GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	serverConfPath := cfg.WireGuard.ServerConfigPath()
	raw, err := os.ReadFile(serverConfPath)
	if err != nil {
		return nil, errors.WithCode(code.ErrWGServerConfigNotFound, "failed to read %s: %v", serverConfPath, err)
	}
	ifaceCfg := wgfile.ParseInterfaceConfig(string(raw))
	if strings.TrimSpace(ifaceCfg.PrivateKey) == "" {
		return nil, errors.WithCode(code.ErrWGServerPrivateKeyMissing, "")
	}

	serverPub, err := wgPubKey(ctx, ifaceCfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	// Parse server Address to get server IP (for exclusion during allocation)
	_, serverIP, err := iputil.ParseFirstV4Prefix(ifaceCfg.Address)
	if err != nil {
		return nil, errors.WithCode(code.ErrWGServerAddressInvalid, "invalid server interface address: %v", err)
	}

	// Extract client IP allocation pool from AllowedIPs in [Peer] blocks
	// AllowedIPs represents networks that clients can access, and we allocate client IPs from the first IPv4 subnet
	allowedIPsList := wgfile.ExtractAllowedIPs(string(raw))
	if len(allowedIPsList) == 0 {
		return nil, errors.WithCode(code.ErrWGAllowedIPsNotFound, "no AllowedIPs found in server config. At least one AllowedIPs entry is required for client IP allocation")
	}

	// Use the first AllowedIPs entry (comma-separated list) to find allocation pool
	var allocationPrefix netip.Prefix
	for _, allowedIPsRaw := range allowedIPsList {
		prefix, err := ipalloc.ParseFirstV4PrefixFromAllowedIPs(allowedIPsRaw)
		if err != nil {
			continue
		}
		allocationPrefix = prefix
		break
	}

	if allocationPrefix == (netip.Prefix{}) {
		return nil, errors.WithCode(code.ErrWGIPv4PrefixNotFound, "no valid IPv4 prefix found in AllowedIPs for client IP allocation")
	}

	// Validate prefix: must have enough bits for allocation
	// /32 = single host (1 IP), /31 = point-to-point (2 IPs), /30 = 4 IPs (2 usable), /29 = 8 IPs (6 usable)
	// We need at least /29 to have enough IPs for clients
	if allocationPrefix.Bits() >= 30 {
		maxHosts := 1 << (32 - allocationPrefix.Bits())
		usableHosts := maxHosts - 2 // subtract network and broadcast
		if allocationPrefix.Bits() == 32 {
			usableHosts = 0 // /32 has no usable hosts for allocation
		}
		return nil, errors.WithCode(code.ErrWGPrefixTooSmall,
			"AllowedIPs prefix too small (%s): only %d usable host(s), need at least /29 (e.g., 100.100.100.0/24) for client IP allocation",
			allocationPrefix.String(), usableHosts)
	}

	// Collect used IPs from database
	peers, _, err := w.store.WGPeers().List(ctx, store.WGPeerListOptions{Limit: 10000})
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
	if req.PrivateKey != nil && strings.TrimSpace(*req.PrivateKey) != "" {
		clientPriv = strings.TrimSpace(*req.PrivateKey)
		// Validate the private key by trying to generate public key from it
		if _, err := wgPubKey(ctx, clientPriv); err != nil {
			return nil, errors.WithCode(code.ErrWGPrivateKeyInvalid, "invalid private key: %v", err)
		}
	} else {
		// Auto-generate private key
		clientPriv, err = wgGenKey(ctx)
		if err != nil {
			return nil, err
		}
	}

	clientPub, err := wgPubKey(ctx, clientPriv)
	if err != nil {
		return nil, err
	}

	peerID, err := snowflake.GenerateID()
	if err != nil {
		return nil, errors.WithCode(code.ErrWGPeerIDGenerationFailed, "failed to generate id: %v", err)
	}

	keepalive := 0
	if req.PersistentKeepalive != nil {
		keepalive = *req.PersistentKeepalive
	}

	dns := ""
	if req.DNS != nil {
		dns = strings.TrimSpace(*req.DNS)
	}
	peer := &model.WGPeer{
		ID:                  peerID,
		UserID:              owner.ID,
		DeviceName:          req.DeviceName,
		ClientPublicKey:     clientPub,
		ClientIP:            clientCIDR,
		AllowedIPs:          strings.TrimSpace(req.AllowedIPs),
		DNS:                 dns,
		PersistentKeepalive: keepalive,
		Status:              model.WGPeerStatusActive,
	}

	// 1) Write user files first (derived artifacts)
	if err := w.writeUserFiles(ctx, owner.Username, peer, clientPriv, serverPub, ifaceCfg.MTU, endpoint); err != nil {
		return nil, err
	}

	// 2) Store peer (source of truth)
	if err := w.store.WGPeers().Create(ctx, peer); err != nil {
		// best-effort cleanup of derived artifacts
		_ = os.RemoveAll(filepath.Join(cfg.WireGuard.ResolvedUserDir(), owner.Username, peer.ID))
		return nil, err
	}

	// 3) Apply server config (lock already held)
	if err := w.syncServerConfigUnlocked(ctx); err != nil {
		return nil, err
	}

	return peer, nil
}

func (w *wgSrv) UserDownloadConfig(ctx context.Context, userID, peerID string) (string, []byte, error) {
	peer, err := w.store.WGPeers().Get(ctx, peerID)
	if err != nil {
		return "", nil, err
	}
	if peer.UserID != userID {
		return "", nil, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied))
	}
	user, err := w.store.Users().GetUser(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	cfg := config.Get()
	confPath := filepath.Join(cfg.WireGuard.ResolvedUserDir(), user.Username, peer.ID, "peer.conf")
	b, err := os.ReadFile(confPath)
	if err != nil {
		return "", nil, errors.WithCode(code.ErrWGUserConfigNotFound, "config not found")
	}
	filename := sanitizeFilename(peer.DeviceName)
	if filename == "" {
		filename = peer.ID
	}
	return filename + ".conf", b, nil
}

func (w *wgSrv) UserRotateConfig(ctx context.Context, userID, peerID string) error {
	peer, err := w.store.WGPeers().Get(ctx, peerID)
	if err != nil {
		return err
	}
	if peer.UserID != userID {
		return errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied))
	}
	cfg := config.Get()

	lockPath := filepath.Join(cfg.WireGuard.RootDir, ".nexuspointwg.lock")
	lock, err := wgfile.AcquireFileLock(lockPath)
	if err != nil {
		return errors.WithCode(code.ErrWGLockAcquireFailed, "failed to acquire wireguard lock: %v", err)
	}
	defer func() { _ = lock.Release() }()

	serverConfPath := cfg.WireGuard.ServerConfigPath()
	raw, err := os.ReadFile(serverConfPath)
	if err != nil {
		return errors.WithCode(code.ErrWGServerConfigNotFound, "failed to read %s: %v", serverConfPath, err)
	}
	ifaceCfg := wgfile.ParseInterfaceConfig(string(raw))
	serverPub, err := wgPubKey(ctx, ifaceCfg.PrivateKey)
	if err != nil {
		return err
	}

	clientPriv, err := wgGenKey(ctx)
	if err != nil {
		return err
	}
	clientPub, err := wgPubKey(ctx, clientPriv)
	if err != nil {
		return err
	}

	user, err := w.store.Users().GetUser(ctx, userID)
	if err != nil {
		return err
	}

	peer.ClientPublicKey = clientPub
	if err := w.store.WGPeers().Update(ctx, peer); err != nil {
		return err
	}
	if err := w.writeUserFiles(ctx, user.Username, peer, clientPriv, serverPub, ifaceCfg.MTU, cfg.WireGuard.Endpoint); err != nil {
		return err
	}
	// Lock already held, use unlocked version
	return w.syncServerConfigUnlocked(ctx)
}

func (w *wgSrv) UserUpdateConfig(ctx context.Context, userID, peerID string, req v1.UserUpdateConfigRequest) error {
	peer, err := w.store.WGPeers().Get(ctx, peerID)
	if err != nil {
		return err
	}
	if peer.UserID != userID {
		return errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied))
	}

	cfg := config.Get()
	lockPath := filepath.Join(cfg.WireGuard.RootDir, ".nexuspointwg.lock")
	lock, err := wgfile.AcquireFileLock(lockPath)
	if err != nil {
		return errors.WithCode(code.ErrWGLockAcquireFailed, "failed to acquire wireguard lock: %v", err)
	}
	defer func() { _ = lock.Release() }()

	// Track if user files need regeneration
	regenerateUserFiles := false

	// Update allowed fields only
	if req.AllowedIPs != nil {
		newAllowedIPs := strings.TrimSpace(*req.AllowedIPs)
		if peer.AllowedIPs != newAllowedIPs {
			peer.AllowedIPs = newAllowedIPs
			regenerateUserFiles = true
		}
	}
	if req.PersistentKeepalive != nil {
		if peer.PersistentKeepalive != *req.PersistentKeepalive {
			peer.PersistentKeepalive = *req.PersistentKeepalive
			regenerateUserFiles = true
		}
	}
	if req.DNS != nil {
		newDNS := strings.TrimSpace(*req.DNS)
		if peer.DNS != newDNS {
			peer.DNS = newDNS
			regenerateUserFiles = true
		}
	}

	// Save to database
	if err := w.store.WGPeers().Update(ctx, peer); err != nil {
		return err
	}

	// Regenerate user files if needed
	if regenerateUserFiles {
		user, uErr := w.store.Users().GetUser(ctx, userID)
		if uErr != nil {
			return uErr
		}

		// Read server config to get server public key and MTU
		serverConfPath := cfg.WireGuard.ServerConfigPath()
		raw, rErr := os.ReadFile(serverConfPath)
		if rErr != nil {
			return errors.WithCode(code.ErrWGServerConfigNotFound, "failed to read %s: %v", serverConfPath, rErr)
		}
		ifaceCfg := wgfile.ParseInterfaceConfig(string(raw))
		serverPub, sErr := wgPubKey(ctx, ifaceCfg.PrivateKey)
		if sErr != nil {
			return sErr
		}

		// Read client private key from file
		baseDir := filepath.Join(cfg.WireGuard.ResolvedUserDir(), user.Username, peer.ID)
		privKeyPath := filepath.Join(baseDir, "privatekey")
		clientPrivBytes, pErr := os.ReadFile(privKeyPath)
		if pErr != nil {
			return errors.WithCode(code.ErrWGPrivateKeyReadFailed, "failed to read private key: %v", pErr)
		}
		clientPriv := strings.TrimSpace(string(clientPrivBytes))

		// Determine endpoint: use request endpoint if provided, otherwise use config default
		endpoint := cfg.WireGuard.Endpoint
		if req.Endpoint != nil && strings.TrimSpace(*req.Endpoint) != "" {
			endpoint = strings.TrimSpace(*req.Endpoint)
		}

		if err := w.writeUserFiles(ctx, user.Username, peer, clientPriv, serverPub, ifaceCfg.MTU, endpoint); err != nil {
			return err
		}
	}

	// Sync server config (lock already held, use unlocked version)
	if err := w.syncServerConfigUnlocked(ctx); err != nil {
		return err
	}

	return nil
}

func (w *wgSrv) UserRevokeConfig(ctx context.Context, userID, peerID string) error {
	peer, err := w.store.WGPeers().Get(ctx, peerID)
	if err != nil {
		return err
	}
	if peer.UserID != userID {
		return errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied))
	}

	cfg := config.Get()
	lockPath := filepath.Join(cfg.WireGuard.RootDir, ".nexuspointwg.lock")
	lock, err := wgfile.AcquireFileLock(lockPath)
	if err != nil {
		return errors.WithCode(code.ErrWGLockAcquireFailed, "failed to acquire wireguard lock: %v", err)
	}
	defer func() { _ = lock.Release() }()

	// Delete database record (hard delete)
	if err := w.store.WGPeers().Delete(ctx, peerID); err != nil {
		return err
	}

	// Delete configuration files (hard delete)
	user, uErr := w.store.Users().GetUser(ctx, userID)
	if uErr == nil {
		_ = os.RemoveAll(filepath.Join(cfg.WireGuard.ResolvedUserDir(), user.Username, peerID))
	}

	// Sync server config to remove peer from server configuration
	// Lock already held, use unlocked version
	if err := w.syncServerConfigUnlocked(ctx); err != nil {
		return err
	}

	return nil
}

func (w *wgSrv) writeUserFiles(ctx context.Context, username string, peer *model.WGPeer, clientPriv string, serverPub string, mtu string, endpoint string) error {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return errors.WithCode(code.ErrWGConfigNotInitialized, "")
	}

	baseDir := filepath.Join(cfg.WireGuard.ResolvedUserDir(), username, peer.ID)
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return errors.WithCode(code.ErrWGUserDirCreateFailed, "failed to create user dir: %v", err)
	}

	// Keys
	if err := os.WriteFile(filepath.Join(baseDir, "privatekey"), []byte(strings.TrimSpace(clientPriv)+"\n"), 0600); err != nil {
		return errors.WithCode(code.ErrWGPrivateKeyWriteFailed, "failed to write private key: %v", err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "publickey"), []byte(strings.TrimSpace(peer.ClientPublicKey)+"\n"), 0600); err != nil {
		return errors.WithCode(code.ErrWGPublicKeyWriteFailed, "failed to write public key: %v", err)
	}

	// Client config
	// 优先使用 peer.AllowedIPs（用户在前端填写的值），如果为空则使用默认值
	clientAllowedIPs := strings.TrimSpace(peer.AllowedIPs)
	if clientAllowedIPs == "" {
		clientAllowedIPs = strings.TrimSpace(cfg.WireGuard.DefaultAllowedIPs)
		if clientAllowedIPs == "" {
			clientAllowedIPs = "0.0.0.0/0,::/0" // 最终默认值
		}
	}
	// 优先使用 peer.DNS（用户在前端填写的值），如果为空则使用默认值
	dns := strings.TrimSpace(peer.DNS)
	if dns == "" {
		dns = strings.TrimSpace(cfg.WireGuard.DNS)
	}
	conf := renderClientConfig(endpoint, dns, clientAllowedIPs, serverPub, peer.ClientIP, clientPriv, peer.PersistentKeepalive, mtu)
	if err := os.WriteFile(filepath.Join(baseDir, "peer.conf"), []byte(conf), 0600); err != nil {
		return errors.WithCode(code.ErrWGConfigWriteFailed, "failed to write peer.conf: %v", err)
	}

	// meta.json (minimal)
	meta := fmt.Sprintf("{\"peer_id\":\"%s\",\"user\":\"%s\",\"device_name\":\"%s\",\"client_ip\":\"%s\",\"endpoint\":\"%s\",\"generated_at\":\"%s\"}\n",
		peer.ID, username, peer.DeviceName, peer.ClientIP, endpoint, time.Now().Format(time.RFC3339))
	_ = os.WriteFile(filepath.Join(baseDir, "meta.json"), []byte(meta), 0600)

	return nil
}

func renderClientConfig(endpoint, dns, defaultAllowed, serverPub, clientCIDR, clientPriv string, keepalive int, mtu string) string {
	var b strings.Builder
	b.WriteString("[Interface]\n")
	b.WriteString("PrivateKey = ")
	b.WriteString(strings.TrimSpace(clientPriv))
	b.WriteString("\n")
	b.WriteString("Address = ")
	b.WriteString(strings.TrimSpace(clientCIDR))
	b.WriteString("\n")
	if strings.TrimSpace(mtu) != "" {
		b.WriteString("MTU = ")
		b.WriteString(strings.TrimSpace(mtu))
		b.WriteString("\n")
	}
	if strings.TrimSpace(dns) != "" {
		b.WriteString("DNS = ")
		b.WriteString(strings.TrimSpace(dns))
		b.WriteString("\n")
	}
	b.WriteString("\n[Peer]\n")
	b.WriteString("PublicKey = ")
	b.WriteString(strings.TrimSpace(serverPub))
	b.WriteString("\n")
	allowed := strings.TrimSpace(defaultAllowed)
	if allowed == "" {
		allowed = "0.0.0.0/0,::/0" // 默认值：允许所有 IPv4 和 IPv6 流量
	}
	b.WriteString("AllowedIPs = ")
	b.WriteString(allowed)
	b.WriteString("\n")
	b.WriteString("Endpoint = ")
	b.WriteString(strings.TrimSpace(endpoint))
	b.WriteString("\n")
	if keepalive > 0 {
		b.WriteString("PersistentKeepalive = ")
		b.WriteString(strconv.Itoa(keepalive))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}

func wgGenKey(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "wg", "genkey")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.WithCode(code.ErrWGKeyGenerationFailed, "wg genkey failed: %s", strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func wgPubKey(ctx context.Context, privateKey string) (string, error) {
	cmd := exec.CommandContext(ctx, "wg", "pubkey")
	cmd.Stdin = bytes.NewBufferString(strings.TrimSpace(privateKey) + "\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.WithCode(code.ErrWGPublicKeyGenerationFailed, "wg pubkey failed: %s", strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

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
