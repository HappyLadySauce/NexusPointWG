package service

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
)

type WGPeerSrv interface {
	CreatePeer(ctx context.Context, peer *model.WGPeer) error
	GetPeer(ctx context.Context, id string) (*model.WGPeer, error)
	ListPeers(ctx context.Context) ([]*model.WGPeer, error)
	ListPeersByUser(ctx context.Context, userID string) ([]*model.WGPeer, error)
	UpdatePeer(ctx context.Context, peer *model.WGPeer) error
	DeletePeer(ctx context.Context, id string) error
}

type wgPeers struct {
	srv *service
}

func newWGPeers(s *service) *wgPeers {
	return &wgPeers{srv: s}
}

func (w *wgPeers) CreatePeer(ctx context.Context, peer *model.WGPeer) error {
	return w.srv.store.WGPeers().CreatePeer(ctx, peer)
}

func (w *wgPeers) GetPeer(ctx context.Context, id string) (*model.WGPeer, error) {
	return w.srv.store.WGPeers().GetPeer(ctx, id)
}

func (w *wgPeers) ListPeers(ctx context.Context) ([]*model.WGPeer, error) {
	return w.srv.store.WGPeers().ListPeers(ctx)
}

func (w *wgPeers) ListPeersByUser(ctx context.Context, userID string) ([]*model.WGPeer, error) {
	return w.srv.store.WGPeers().ListPeersByUser(ctx, userID)
}

func (w *wgPeers) UpdatePeer(ctx context.Context, peer *model.WGPeer) error {
	return w.srv.store.WGPeers().UpdatePeer(ctx, peer)
}

func (w *wgPeers) DeletePeer(ctx context.Context, id string) error {
	return w.srv.store.WGPeers().DeletePeer(ctx, id)
}
