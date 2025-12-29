package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	wgfile "github.com/HappyLadySauce/NexusPointWG/internal/pkg/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/errors"
	"k8s.io/klog/v2"
)

type WGSrv interface {
	// SyncServerConfig rewrites wg0.conf managed block from DB and applies it (if enabled).
	SyncServerConfig(ctx context.Context) error

	// ---- admin peer ops ----
	AdminListPeers(ctx context.Context, opt store.WGPeerListOptions) ([]*model.WGPeer, int64, error)
	AdminCreatePeer(ctx context.Context, req v1.CreateWGPeerRequest) (*model.WGPeer, error)
	AdminGetPeer(ctx context.Context, id string) (*model.WGPeer, error)
	AdminUpdatePeer(ctx context.Context, id string, req v1.UpdateWGPeerRequest) (*model.WGPeer, error)
	AdminRevokePeer(ctx context.Context, id string) error

	// ---- user config ops ----
	UserListPeers(ctx context.Context, userID string) ([]*model.WGPeer, error)
	UserDownloadConfig(ctx context.Context, userID, peerID string) (filename string, content []byte, err error)
	UserRotateConfig(ctx context.Context, userID, peerID string) error
	UserRevokeConfig(ctx context.Context, userID, peerID string) error

	// ---- response mappers ----
	ToWGPeerResponse(ctx context.Context, peer *model.WGPeer) (*v1.WGPeerResponse, error)
	ToWGPeerListResponse(ctx context.Context, peers []*model.WGPeer, total int64) (*v1.WGPeerListResponse, error)
}

type wgSrv struct {
	storeSvc *service
}

var _ WGSrv = (*wgSrv)(nil)

func newWG(s *service) *wgSrv {
	return &wgSrv{storeSvc: s}
}

func (w *wgSrv) SyncServerConfig(ctx context.Context) error {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return errors.WithCode(code.ErrUnknown, "wireguard config is not initialized")
	}

	rootDir := cfg.WireGuard.RootDir
	lockPath := filepath.Join(rootDir, ".nexuspointwg.lock")

	lock, err := wgfile.AcquireFileLock(lockPath)
	if err != nil {
		return errors.WithCode(code.ErrUnknown, "failed to acquire wireguard lock: %v", err)
	}
	defer func() { _ = lock.Release() }()

	return w.syncServerConfigUnlocked(ctx)
}

// syncServerConfigUnlocked performs the actual server config sync without acquiring a lock.
// It should only be called when the caller already holds the lock.
func (w *wgSrv) syncServerConfigUnlocked(ctx context.Context) error {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return errors.WithCode(code.ErrUnknown, "wireguard config is not initialized")
	}

	peers, err := w.loadAllPeers(ctx)
	if err != nil {
		return err
	}

	serverConf := cfg.WireGuard.ServerConfigPath()
	orig, err := os.ReadFile(serverConf)
	if err != nil {
		return errors.WithCode(code.ErrWGServerConfigNotFound, "failed to read %s: %v", serverConf, err)
	}

	managed := wgfile.RenderManagedBlock(peers)
	updated := wgfile.ReplaceOrAppendManagedBlock(string(orig), managed)

	// Preserve existing file perms where possible.
	var perm os.FileMode = 0600
	if st, statErr := os.Stat(serverConf); statErr == nil {
		perm = st.Mode().Perm()
	}
	if err := wgfile.AtomicWriteFile(serverConf, []byte(updated), perm); err != nil {
		return errors.WithCode(code.ErrWGWriteServerConfigFailed, "failed to write %s: %v", serverConf, err)
	}

	switch strings.ToLower(strings.TrimSpace(cfg.WireGuard.ApplyMethod)) {
	case "", "systemctl":
		return w.systemctlRestart(ctx, cfg.WireGuard.Interface)
	case "none":
		return nil
	default:
		// Validate() should already prevent this.
		return errors.WithCode(code.ErrBind, "invalid wireguard.apply-method")
	}
}

func (w *wgSrv) loadAllPeers(ctx context.Context) ([]*model.WGPeer, error) {
	// Load all peers; filter revoked at render stage.
	peers, _, err := w.storeSvc.store.WGPeers().List(ctx, store.WGPeerListOptions{
		Offset: 0,
		Limit:  10000,
	})
	if err != nil {
		return nil, err
	}
	return peers, nil
}

func (w *wgSrv) systemctlRestart(ctx context.Context, iface string) error {
	if strings.TrimSpace(iface) == "" {
		return errors.WithCode(code.ErrBind, "wireguard.interface is required")
	}

	unit := fmt.Sprintf("wg-quick@%s", iface)

	// Keep apply timeout bounded; systemd can hang under some failure modes.
	applyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(applyCtx, "systemctl", "restart", unit)
	out, err := cmd.CombinedOutput()
	if err != nil {
		klog.V(1).InfoS("systemctl restart failed", "unit", unit, "output", string(out), "error", err)
		return errors.WithCode(code.ErrWGApplyFailed, "systemctl restart %s failed: %s", unit, strings.TrimSpace(string(out)))
	}
	return nil
}
