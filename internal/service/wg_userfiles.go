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
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	wgfile "github.com/HappyLadySauce/NexusPointWG/internal/pkg/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/snowflake"
	"github.com/HappyLadySauce/errors"
)

func (w *wgSrv) AdminCreatePeer(ctx context.Context, req v1.CreateWGPeerRequest) (*model.WGPeer, error) {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return nil, errors.WithCode(code.ErrUnknown, "wireguard config is not initialized")
	}
	if strings.TrimSpace(cfg.WireGuard.Endpoint) == "" {
		return nil, errors.WithCode(code.ErrValidation, "wireguard.endpoint is required")
	}

	// Serialize allocate-ip + file write + server apply.
	lockPath := filepath.Join(cfg.WireGuard.RootDir, ".nexuspointwg.lock")
	lock, err := wgfile.AcquireFileLock(lockPath)
	if err != nil {
		return nil, errors.WithCode(code.ErrUnknown, "failed to acquire wireguard lock: %v", err)
	}
	defer func() { _ = lock.Release() }()

	owner, err := w.storeSvc.store.Users().GetUserByUsername(ctx, req.Username)
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
		return nil, errors.WithCode(code.ErrValidation, "server config missing Interface.PrivateKey")
	}

	serverPub, err := wgPubKey(ctx, ifaceCfg.PrivateKey)
	if err != nil {
		return nil, err
	}

	prefix, serverIP, err := parseFirstV4Prefix(ifaceCfg.Address)
	if err != nil {
		return nil, errors.WithCode(code.ErrValidation, "invalid server interface address: %v", err)
	}

	used, err := w.collectUsedIPs(ctx, prefix, string(raw))
	if err != nil {
		return nil, err
	}

	clientAddr, err := allocateIPv4(prefix, serverIP, used)
	if err != nil {
		return nil, errors.WithCode(code.ErrValidation, "no available ip in %s", prefix.String())
	}
	clientCIDR := clientAddr.String() + "/32"

	clientPriv, err := wgGenKey(ctx)
	if err != nil {
		return nil, err
	}
	clientPub, err := wgPubKey(ctx, clientPriv)
	if err != nil {
		return nil, err
	}

	peerID, err := snowflake.GenerateID()
	if err != nil {
		return nil, errors.WithCode(code.ErrUnknown, "failed to generate id: %v", err)
	}

	keepalive := 0
	if req.PersistentKeepalive != nil {
		keepalive = *req.PersistentKeepalive
	}

	peer := &model.WGPeer{
		ID:                  peerID,
		UserID:              owner.ID,
		DeviceName:          req.DeviceName,
		ClientPublicKey:     clientPub,
		ClientIP:            clientCIDR,
		AllowedIPs:          strings.TrimSpace(req.AllowedIPs),
		PersistentKeepalive: keepalive,
		Status:              model.WGPeerStatusActive,
	}

	// 1) Write user files first (derived artifacts)
	if err := w.writeUserFiles(ctx, owner.Username, peer, clientPriv, serverPub, ifaceCfg.MTU); err != nil {
		return nil, err
	}

	// 2) Store peer (source of truth)
	if err := w.storeSvc.store.WGPeers().Create(ctx, peer); err != nil {
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
	peer, err := w.storeSvc.store.WGPeers().Get(ctx, peerID)
	if err != nil {
		return "", nil, err
	}
	if peer.UserID != userID {
		return "", nil, errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied))
	}
	user, err := w.storeSvc.store.Users().GetUser(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	cfg := config.Get()
	confPath := filepath.Join(cfg.WireGuard.ResolvedUserDir(), user.Username, peer.ID, "peer.conf")
	b, err := os.ReadFile(confPath)
	if err != nil {
		return "", nil, errors.WithCode(code.ErrWGServerConfigNotFound, "config not found")
	}
	filename := sanitizeFilename(peer.DeviceName)
	if filename == "" {
		filename = peer.ID
	}
	return filename + ".conf", b, nil
}

func (w *wgSrv) UserRotateConfig(ctx context.Context, userID, peerID string) error {
	peer, err := w.storeSvc.store.WGPeers().Get(ctx, peerID)
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
		return errors.WithCode(code.ErrUnknown, "failed to acquire wireguard lock: %v", err)
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

	user, err := w.storeSvc.store.Users().GetUser(ctx, userID)
	if err != nil {
		return err
	}

	peer.ClientPublicKey = clientPub
	if err := w.storeSvc.store.WGPeers().Update(ctx, peer); err != nil {
		return err
	}
	if err := w.writeUserFiles(ctx, user.Username, peer, clientPriv, serverPub, ifaceCfg.MTU); err != nil {
		return err
	}
	// Lock already held, use unlocked version
	return w.syncServerConfigUnlocked(ctx)
}

func (w *wgSrv) UserRevokeConfig(ctx context.Context, userID, peerID string) error {
	peer, err := w.storeSvc.store.WGPeers().Get(ctx, peerID)
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
		return errors.WithCode(code.ErrUnknown, "failed to acquire wireguard lock: %v", err)
	}
	defer func() { _ = lock.Release() }()

	peer.Status = model.WGPeerStatusRevoked
	if err := w.storeSvc.store.WGPeers().Update(ctx, peer); err != nil {
		return err
	}
	// Lock already held, use unlocked version
	if err := w.syncServerConfigUnlocked(ctx); err != nil {
		return err
	}
	// Best-effort cleanup of derived artifacts.
	user, uErr := w.storeSvc.store.Users().GetUser(ctx, userID)
	if uErr == nil {
		_ = os.RemoveAll(filepath.Join(cfg.WireGuard.ResolvedUserDir(), user.Username, peer.ID))
	}
	return nil
}

func (w *wgSrv) writeUserFiles(ctx context.Context, username string, peer *model.WGPeer, clientPriv string, serverPub string, mtu string) error {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return errors.WithCode(code.ErrUnknown, "wireguard config is not initialized")
	}

	baseDir := filepath.Join(cfg.WireGuard.ResolvedUserDir(), username, peer.ID)
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return errors.WithCode(code.ErrUnknown, "failed to create user dir: %v", err)
	}

	// Keys
	if err := os.WriteFile(filepath.Join(baseDir, "privatekey"), []byte(strings.TrimSpace(clientPriv)+"\n"), 0600); err != nil {
		return errors.WithCode(code.ErrUnknown, "failed to write private key: %v", err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "publickey"), []byte(strings.TrimSpace(peer.ClientPublicKey)+"\n"), 0600); err != nil {
		return errors.WithCode(code.ErrUnknown, "failed to write public key: %v", err)
	}

	// Client config
	conf := renderClientConfig(cfg.WireGuard.Endpoint, cfg.WireGuard.DNS, cfg.WireGuard.DefaultAllowedIPs, serverPub, peer.ClientIP, clientPriv, peer.PersistentKeepalive, mtu)
	if err := os.WriteFile(filepath.Join(baseDir, "peer.conf"), []byte(conf), 0600); err != nil {
		return errors.WithCode(code.ErrUnknown, "failed to write peer.conf: %v", err)
	}

	// meta.json (minimal)
	meta := fmt.Sprintf("{\"peer_id\":\"%s\",\"user\":\"%s\",\"device_name\":\"%s\",\"client_ip\":\"%s\",\"endpoint\":\"%s\",\"generated_at\":\"%s\"}\n",
		peer.ID, username, peer.DeviceName, peer.ClientIP, cfg.WireGuard.Endpoint, time.Now().Format(time.RFC3339))
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
		allowed = "0.0.0.0/0"
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

func parseFirstV4Prefix(addressLine string) (netip.Prefix, netip.Addr, error) {
	// Address may be comma-separated.
	parts := strings.Split(addressLine, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(p)
		if err != nil {
			continue
		}
		if prefix.Addr().Is4() {
			return prefix.Masked(), prefix.Addr(), nil
		}
	}
	return netip.Prefix{}, netip.Addr{}, fmt.Errorf("no ipv4 prefix found")
}

func (w *wgSrv) collectUsedIPs(ctx context.Context, prefix netip.Prefix, serverConf string) (map[netip.Addr]struct{}, error) {
	used := make(map[netip.Addr]struct{})

	// From DB
	peers, _, err := w.storeSvc.store.WGPeers().List(ctx, store.WGPeerListOptions{Limit: 10000})
	if err != nil {
		return nil, err
	}
	for _, p := range peers {
		if p == nil {
			continue
		}
		cidr := strings.TrimSpace(p.ClientIP)
		if cidr == "" {
			continue
		}
		pr, err := netip.ParsePrefix(cidr)
		if err != nil {
			continue
		}
		if pr.Addr().Is4() && prefix.Contains(pr.Addr()) {
			used[pr.Addr()] = struct{}{}
		}
	}

	// From existing config file (including non-managed blocks)
	for _, raw := range wgfile.ExtractAllowedIPs(serverConf) {
		for _, cidr := range splitCSV(raw) {
			pr, err := netip.ParsePrefix(cidr)
			if err != nil {
				continue
			}
			if pr.Addr().Is4() && prefix.Contains(pr.Addr()) {
				used[pr.Addr()] = struct{}{}
			}
		}
	}

	return used, nil
}

func allocateIPv4(prefix netip.Prefix, serverIP netip.Addr, used map[netip.Addr]struct{}) (netip.Addr, error) {
	// Iterate hosts in prefix. For typical /24 this is fine.
	start := prefix.Masked().Addr()
	last := lastIPv4(prefix)
	for ip := start; prefix.Contains(ip); ip = ip.Next() {
		if !ip.Is4() {
			continue
		}
		// skip network addr
		if ip == start {
			continue
		}
		// skip broadcast addr
		if ip == last {
			continue
		}
		// skip server ip
		if ip == serverIP {
			continue
		}
		if _, ok := used[ip]; ok {
			continue
		}
		return ip, nil
	}
	return netip.Addr{}, fmt.Errorf("no available ip")
}

func lastIPv4(prefix netip.Prefix) netip.Addr {
	p := prefix.Masked()
	if !p.Addr().Is4() {
		return netip.Addr{}
	}
	base := p.Addr().As4()
	ones := p.Bits()
	hostBits := 32 - ones
	var n uint32
	n |= uint32(base[0]) << 24
	n |= uint32(base[1]) << 16
	n |= uint32(base[2]) << 8
	n |= uint32(base[3])
	if hostBits >= 32 {
		n |= ^uint32(0)
	} else if hostBits > 0 {
		n |= (uint32(1) << hostBits) - 1
	}
	b0 := byte(n >> 24)
	b1 := byte(n >> 16)
	b2 := byte(n >> 8)
	b3 := byte(n)
	return netip.AddrFrom4([4]byte{b0, b1, b2, b3})
}

func wgGenKey(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "wg", "genkey")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.WithCode(code.ErrUnknown, "wg genkey failed: %s", strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func wgPubKey(ctx context.Context, privateKey string) (string, error) {
	cmd := exec.CommandContext(ctx, "wg", "pubkey")
	cmd.Stdin = bytes.NewBufferString(strings.TrimSpace(privateKey) + "\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.WithCode(code.ErrUnknown, "wg pubkey failed: %s", strings.TrimSpace(string(out)))
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
