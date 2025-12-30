package service

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/HappyLadySauce/NexusPointWG/internal/local"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	wgfile "github.com/HappyLadySauce/NexusPointWG/internal/pkg/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/errors"
	"k8s.io/klog/v2"
)

// CreatePeerParams represents parameters for creating a peer.
type CreatePeerParams struct {
	Username            string
	DeviceName          string
	AllowedIPs          string
	PersistentKeepalive *int
	Endpoint            *string
	DNS                 *string
	PrivateKey          *string
}

// UpdatePeerParams represents parameters for updating a peer.
type UpdatePeerParams struct {
	AllowedIPs          *string
	PersistentKeepalive *int
	DNS                 *string
	Status              *string
	PrivateKey          *string
	Endpoint            *string
	DeviceName          *string
	ClientIP            *string
}

type WGSrv interface {
	// SyncServerConfig rewrites wg0.conf managed block from DB and applies it (if enabled).
	SyncServerConfig(ctx context.Context) error

	// ListPeers lists peers with filtering and pagination.
	ListPeers(ctx context.Context, opt store.WGPeerListOptions) ([]*model.WGPeer, int64, error)

	// GetPeer returns a peer by ID.
	GetPeer(ctx context.Context, id string) (*model.WGPeer, error)

	// CreatePeer creates a new peer.
	CreatePeer(ctx context.Context, params CreatePeerParams) (*model.WGPeer, error)

	// UpdatePeer updates a peer.
	UpdatePeer(ctx context.Context, id string, params UpdatePeerParams) (*model.WGPeer, error)

	// DeletePeer deletes a peer.
	DeletePeer(ctx context.Context, id string) error

	// DownloadConfig downloads a peer configuration file.
	DownloadConfig(ctx context.Context, peerID string) (filename string, content []byte, err error)

	// RotateConfig rotates keys for a peer.
	RotateConfig(ctx context.Context, peerID string) error
}

type wgSrv struct {
	store store.Factory
}

var _ WGSrv = (*wgSrv)(nil)

func newWG(s *service) *wgSrv {
	return &wgSrv{store: s.store}
}

func (w *wgSrv) SyncServerConfig(ctx context.Context) error {
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return errors.WithCode(code.ErrUnknown, "wireguard config is not initialized")
	}

	// TODO: File locking will be handled later

	// Load all peers from database
	peers, _, err := w.ListPeers(ctx, store.WGPeerListOptions{
		Offset: 0,
		Limit:  10000,
	})
	if err != nil {
		return err
	}

	serverConfigStore := local.NewLocalStore().ServerConfigStore()
	origBytes, err := serverConfigStore.Read(ctx)
	if err != nil {
		return err
	}

	// Parse the server configuration using structured parsing
	serverConfig, err := wgfile.ParseServerConfig(origBytes)
	if err != nil {
		return errors.WithCode(code.ErrWGServerConfigParseFailed, "failed to parse server config: %v", err)
	}

	// Replace managed block peers using structured approach
	wgfile.ReplaceManagedBlockPeers(serverConfig, peers)

	// Render the updated configuration
	updatedBytes := wgfile.RenderServerConfig(serverConfig)

	if err := serverConfigStore.Write(ctx, updatedBytes); err != nil {
		return err
	}

	// Apply configuration if enabled
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
