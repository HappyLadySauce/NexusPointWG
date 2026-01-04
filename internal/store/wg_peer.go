package store

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
)

// WGPeerStore defines the interface for WireGuard peer data access.
type WGPeerStore interface {
	// CreatePeer creates a new WireGuard peer.
	CreatePeer(ctx context.Context, peer *model.WGPeer) error

	// GetPeer retrieves a peer by ID.
	GetPeer(ctx context.Context, id string) (*model.WGPeer, error)

	// GetPeerByPublicKey retrieves a peer by public key.
	GetPeerByPublicKey(ctx context.Context, publicKey string) (*model.WGPeer, error)

	// UpdatePeer updates an existing peer.
	UpdatePeer(ctx context.Context, peer *model.WGPeer) error

	// DeletePeer deletes a peer by ID.
	DeletePeer(ctx context.Context, id string) error

	// ListPeers lists peers with optional filters and pagination.
	ListPeers(ctx context.Context, opt WGPeerListOptions) ([]*model.WGPeer, int64, error)

	// CountPeersByUserID counts the number of peers for a specific user.
	CountPeersByUserID(ctx context.Context, userID string) (int64, error)
}

// WGPeerListOptions defines options for listing WireGuard peers.
type WGPeerListOptions struct {
	UserID     string
	Status     string
	IPPoolID   string
	DeviceName string
	Offset     int
	Limit      int
}

