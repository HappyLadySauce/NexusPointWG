package service

import (
	"context"
	"os"
	"path/filepath"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	v1 "github.com/HappyLadySauce/NexusPointWG/internal/pkg/types/v1"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/errors"
)

func (w *wgSrv) AdminListPeers(ctx context.Context, opt store.WGPeerListOptions) ([]*model.WGPeer, int64, error) {
	return w.storeSvc.store.WGPeers().List(ctx, opt)
}

func (w *wgSrv) AdminGetPeer(ctx context.Context, id string) (*model.WGPeer, error) {
	return w.storeSvc.store.WGPeers().Get(ctx, id)
}

func (w *wgSrv) AdminUpdatePeer(ctx context.Context, id string, req v1.UpdateWGPeerRequest) (*model.WGPeer, error) {
	peer, err := w.storeSvc.store.WGPeers().Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.AllowedIPs != nil {
		peer.AllowedIPs = *req.AllowedIPs
	}
	if req.PersistentKeepalive != nil {
		peer.PersistentKeepalive = *req.PersistentKeepalive
	}
	if req.Status != nil {
		peer.Status = *req.Status
	}
	if err := w.storeSvc.store.WGPeers().Update(ctx, peer); err != nil {
		return nil, err
	}
	// Apply changes to server config (best effort, but fail fast if apply fails).
	if err := w.SyncServerConfig(ctx); err != nil {
		return nil, err
	}
	return peer, nil
}

func (w *wgSrv) AdminRevokePeer(ctx context.Context, id string) error {
	peer, err := w.storeSvc.store.WGPeers().Get(ctx, id)
	if err != nil {
		return err
	}
	peer.Status = model.WGPeerStatusRevoked
	if err := w.storeSvc.store.WGPeers().Update(ctx, peer); err != nil {
		return err
	}
	if err := w.SyncServerConfig(ctx); err != nil {
		return err
	}
	// Best-effort cleanup of derived artifacts.
	cfg := config.Get()
	if cfg != nil && cfg.WireGuard != nil {
		user, uErr := w.storeSvc.store.Users().GetUser(ctx, peer.UserID)
		if uErr == nil {
			_ = os.RemoveAll(filepath.Join(cfg.WireGuard.ResolvedUserDir(), user.Username, peer.ID))
		}
	}
	return nil
}

func (w *wgSrv) UserListPeers(ctx context.Context, userID string) ([]*model.WGPeer, error) {
	peers, _, err := w.storeSvc.store.WGPeers().List(ctx, store.WGPeerListOptions{
		UserID: userID,
		Offset: 0,
		Limit:  10000,
	})
	return peers, err
}

func (w *wgSrv) ToWGPeerResponse(ctx context.Context, peer *model.WGPeer) (*v1.WGPeerResponse, error) {
	if peer == nil {
		return nil, errors.WithCode(code.ErrValidation, "peer is nil")
	}
	user, err := w.storeSvc.store.Users().GetUser(ctx, peer.UserID)
	if err != nil {
		return nil, err
	}
	return &v1.WGPeerResponse{
		ID:                  peer.ID,
		UserID:              peer.UserID,
		Username:            user.Username,
		DeviceName:          peer.DeviceName,
		ClientPublicKey:     peer.ClientPublicKey,
		ClientIP:            peer.ClientIP,
		AllowedIPs:          peer.AllowedIPs,
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
