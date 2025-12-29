package service

import (
	"context"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	ipalloc "github.com/HappyLadySauce/NexusPointWG/internal/pkg/ipalloc"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	wgfile "github.com/HappyLadySauce/NexusPointWG/internal/pkg/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	iputil "github.com/HappyLadySauce/NexusPointWG/pkg/utils/ip"
	"github.com/HappyLadySauce/errors"
	"k8s.io/klog/v2"
)

func (w *wgSrv) AdminListPeers(ctx context.Context, opt store.WGPeerListOptions) ([]*model.WGPeer, int64, error) {
	return w.store.WGPeers().List(ctx, opt)
}

func (w *wgSrv) AdminGetPeer(ctx context.Context, id string) (*model.WGPeer, error) {
	return w.store.WGPeers().Get(ctx, id)
}

// validateClientIPUpdate 验证 ClientIP 更新是否有效
// 参数：
//   - ctx: 上下文
//   - newClientIP: 新的 ClientIP（CIDR 格式，例如：100.100.100.5/32）
//   - currentPeerID: 当前正在更新的 peer ID（用于排除自己）
//
// 返回：
//   - error: 如果验证失败返回错误
func (w *wgSrv) validateClientIPUpdate(ctx context.Context, newClientIP string, currentPeerID string) error {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return errors.WithCode(code.ErrWGConfigNotInitialized, code.Message(code.ErrWGConfigNotInitialized))
	}

	// 1. 解析新的 ClientIP
	prefix, err := netip.ParsePrefix(newClientIP)
	if err != nil {
		return errors.WithCode(code.ErrValidation, "invalid ClientIP format: %v", err)
	}
	ip := prefix.Addr()
	if !ip.Is4() {
		return errors.WithCode(code.ErrIPNotIPv4, code.Message(code.ErrIPNotIPv4))
	}

	// 2. 读取服务器配置，获取分配池信息
	serverConfPath := cfg.WireGuard.ServerConfigPath()
	raw, err := os.ReadFile(serverConfPath)
	if err != nil {
		return errors.WithCode(code.ErrWGServerConfigNotFound, "failed to read %s: %v", serverConfPath, err)
	}
	ifaceCfg := wgfile.ParseInterfaceConfig(string(raw))

	// 获取服务器 IP
	_, serverIP, err := iputil.ParseFirstV4Prefix(ifaceCfg.Address)
	if err != nil {
		return errors.WithCode(code.ErrWGServerAddressInvalid, "invalid server interface address: %v", err)
	}

	// 从 AllowedIPs 中提取分配前缀
	allowedIPsList := wgfile.ExtractAllowedIPs(string(raw))
	if len(allowedIPsList) == 0 {
		return errors.WithCode(code.ErrWGAllowedIPsNotFound, "")
	}

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
		return errors.WithCode(code.ErrWGIPv4PrefixNotFound, "")
	}

	// 3. 收集已使用的 IP（排除当前 peer）
	peers, _, err := w.store.WGPeers().List(ctx, store.WGPeerListOptions{Limit: 10000})
	if err != nil {
		return err
	}

	// 过滤掉当前 peer
	filteredPeers := make([]*model.WGPeer, 0, len(peers))
	for _, p := range peers {
		if p != nil && p.ID != currentPeerID {
			filteredPeers = append(filteredPeers, p)
		}
	}

	usedIPs := ipalloc.CollectUsedIPsFromPeers(filteredPeers, allocationPrefix)

	// 4. 创建 Allocator 并验证
	allocator := ipalloc.NewAllocator(allocationPrefix, serverIP, usedIPs)
	if err := allocator.Validate(ip); err != nil {
		return err
	}

	return nil
}

func (w *wgSrv) AdminUpdatePeer(ctx context.Context, id string, req v1.UpdateWGPeerRequest) (*model.WGPeer, error) {
	peer, err := w.store.WGPeers().Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Track if user files need regeneration
	regenerateUserFiles := false
	privateKeyChanged := false

	// Update fields
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
	if req.Status != nil {
		peer.Status = *req.Status
	}
	if req.DeviceName != nil {
		peer.DeviceName = strings.TrimSpace(*req.DeviceName)
	}
	if req.PrivateKey != nil {
		newPriv := strings.TrimSpace(*req.PrivateKey)
		// Validate the private key by trying to generate public key from it
		if _, err := wgPubKey(ctx, newPriv); err != nil {
			return nil, errors.WithCode(code.ErrWGPrivateKeyInvalid, "invalid private key: %v", err)
		}
		privateKeyChanged = true
		regenerateUserFiles = true
	}
	if req.ClientIP != nil {
		newClientIP := strings.TrimSpace(*req.ClientIP)
		// Validate ClientIP is available (conflict check)
		if err := w.validateClientIPUpdate(ctx, newClientIP, id); err != nil {
			return nil, err
		}
		if peer.ClientIP != newClientIP {
			peer.ClientIP = newClientIP
			regenerateUserFiles = true
		}
	}

	// If PrivateKey changed, regenerate public key
	if privateKeyChanged {
		newPriv := strings.TrimSpace(*req.PrivateKey)
		newPub, err := wgPubKey(ctx, newPriv)
		if err != nil {
			return nil, err
		}
		peer.ClientPublicKey = newPub
	}

	// Save to database
	if err := w.store.WGPeers().Update(ctx, peer); err != nil {
		return nil, err
	}

	// Regenerate user files if needed
	if regenerateUserFiles {
		cfg := config.Get()
		if cfg != nil && cfg.WireGuard != nil {
			user, uErr := w.store.Users().GetUser(ctx, peer.UserID)
			if uErr == nil {
				// Read server config to get server public key and MTU
				serverConfPath := cfg.WireGuard.ServerConfigPath()
				raw, rErr := os.ReadFile(serverConfPath)
				if rErr == nil {
					ifaceCfg := wgfile.ParseInterfaceConfig(string(raw))
					serverPub, sErr := wgPubKey(ctx, ifaceCfg.PrivateKey)
					if sErr == nil {
						// Read or use new client private key
						var clientPriv string
						if privateKeyChanged {
							clientPriv = strings.TrimSpace(*req.PrivateKey)
						} else {
							baseDir := filepath.Join(cfg.WireGuard.ResolvedUserDir(), user.Username, peer.ID)
							privKeyPath := filepath.Join(baseDir, "privatekey")
							clientPrivBytes, pErr := os.ReadFile(privKeyPath)
							if pErr == nil {
								clientPriv = strings.TrimSpace(string(clientPrivBytes))
							} else {
								klog.V(1).InfoS("failed to read private key", "error", pErr)
								// Continue with endpoint update even if private key read fails
							}
						}

						// Determine endpoint: use request endpoint if provided, otherwise use config default
						endpoint := cfg.WireGuard.Endpoint
						if req.Endpoint != nil && strings.TrimSpace(*req.Endpoint) != "" {
							endpoint = strings.TrimSpace(*req.Endpoint)
						}

						if clientPriv != "" {
							if err := w.writeUserFiles(ctx, user.Username, peer, clientPriv, serverPub, ifaceCfg.MTU, endpoint); err != nil {
								// Log error but don't fail the update
								klog.V(1).InfoS("failed to regenerate user files after update", "error", err)
							}
						}
					}
				}
			}
		}
	}

	// Apply changes to server config (best effort, but fail fast if apply fails).
	if err := w.SyncServerConfig(ctx); err != nil {
		return nil, err
	}
	return peer, nil
}

func (w *wgSrv) AdminRevokePeer(ctx context.Context, id string) error {
	peer, err := w.store.WGPeers().Get(ctx, id)
	if err != nil {
		return err
	}

	// Delete database record (hard delete)
	if err := w.store.WGPeers().Delete(ctx, id); err != nil {
		return err
	}

	// Delete configuration files (hard delete)
	cfg := config.Get()
	if cfg != nil && cfg.WireGuard != nil {
		user, uErr := w.store.Users().GetUser(ctx, peer.UserID)
		if uErr == nil {
			_ = os.RemoveAll(filepath.Join(cfg.WireGuard.ResolvedUserDir(), user.Username, peer.ID))
		}
	}

	// Sync server config to remove peer from server configuration
	if err := w.SyncServerConfig(ctx); err != nil {
		return err
	}

	return nil
}

func (w *wgSrv) UserListPeers(ctx context.Context, userID string) ([]*model.WGPeer, error) {
	peers, _, err := w.store.WGPeers().List(ctx, store.WGPeerListOptions{
		UserID: userID,
		Offset: 0,
		Limit:  10000,
	})
	return peers, err
}

func (w *wgSrv) ToWGPeerResponse(ctx context.Context, peer *model.WGPeer) (*v1.WGPeerResponse, error) {
	if peer == nil {
		return nil, errors.WithCode(code.ErrWGPeerNil, "")
	}
	user, err := w.store.Users().GetUser(ctx, peer.UserID)
	if err != nil {
		return nil, err
	}
	// 如果 AllowedIPs 为空，返回实际使用的默认值
	allowedIPs := peer.AllowedIPs
	if strings.TrimSpace(allowedIPs) == "" {
		// 获取配置中的默认值
		cfg := config.Get()
		if cfg != nil && cfg.WireGuard != nil {
			allowedIPs = strings.TrimSpace(cfg.WireGuard.DefaultAllowedIPs)
			if allowedIPs == "" {
				allowedIPs = "0.0.0.0/0,::/0" // 最终默认值
			}
		} else {
			allowedIPs = "0.0.0.0/0,::/0" // 兜底默认值
		}
	}
	// 如果 DNS 为空，返回实际使用的默认值
	dns := peer.DNS
	if strings.TrimSpace(dns) == "" {
		cfg := config.Get()
		if cfg != nil && cfg.WireGuard != nil {
			dns = strings.TrimSpace(cfg.WireGuard.DNS)
		}
	}
	return &v1.WGPeerResponse{
		ID:                  peer.ID,
		UserID:              peer.UserID,
		Username:            user.Username,
		DeviceName:          peer.DeviceName,
		ClientPublicKey:     peer.ClientPublicKey,
		ClientIP:            peer.ClientIP,
		AllowedIPs:          allowedIPs,
		DNS:                 dns,
		PersistentKeepalive: peer.PersistentKeepalive,
		Status:              peer.Status,
	}, nil
}

func (w *wgSrv) ToWGPeerListResponse(ctx context.Context, peers []*model.WGPeer, total int64) (*v1.WGPeerListResponse, error) {
	items := make([]v1.WGPeerResponse, 0, len(peers))
	for _, p := range peers {
		resp, err := w.ToWGPeerResponse(ctx, p)
		if err != nil {
			return nil, err
		}
		items = append(items, *resp)
	}
	return &v1.WGPeerListResponse{
		Total: total,
		Items: items,
	}, nil
}
