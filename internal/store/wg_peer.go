package store

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
)

// WGPeerStore defines storage operations for WireGuard peers.
type WGPeerStore interface {
	Create(ctx context.Context, peer *model.WGPeer) error
	Get(ctx context.Context, id string) (*model.WGPeer, error)
	Update(ctx context.Context, peer *model.WGPeer) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, opt WGPeerListOptions) ([]*model.WGPeer, int64, error)
}

// WGPeerListOptions defines list filters and pagination.
type WGPeerListOptions struct {
	UserID     string
	DeviceName string
	Status     string

	// Pagination
	Offset int
	Limit  int
}
