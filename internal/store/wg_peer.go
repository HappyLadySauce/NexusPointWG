package store

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
)

type WGPeerStore interface {
	CreatePeer(ctx context.Context, peer *model.WGPeer) error
	GetPeer(ctx context.Context, id string) (*model.WGPeer, error)
	ListPeers(ctx context.Context) ([]*model.WGPeer, error)
	ListPeersByUser(ctx context.Context, userID string) ([]*model.WGPeer, error)
	UpdatePeer(ctx context.Context, peer *model.WGPeer) error
	DeletePeer(ctx context.Context, id string) error
}
