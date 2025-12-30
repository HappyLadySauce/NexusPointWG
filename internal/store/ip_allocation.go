package store

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
)

// IPAllocationStore defines the interface for IP allocation data access.
type IPAllocationStore interface {
	// CreateIPAllocation creates a new IP allocation record.
	CreateIPAllocation(ctx context.Context, allocation *model.IPAllocation) error

	// GetIPAllocation retrieves an IP allocation by ID.
	GetIPAllocation(ctx context.Context, id string) (*model.IPAllocation, error)

	// GetIPAllocationByIPAddress retrieves an IP allocation by IP address.
	GetIPAllocationByIPAddress(ctx context.Context, ipAddress string) (*model.IPAllocation, error)

	// GetIPAllocationByPeerID retrieves an IP allocation by peer ID.
	GetIPAllocationByPeerID(ctx context.Context, peerID string) (*model.IPAllocation, error)

	// UpdateIPAllocation updates an existing IP allocation.
	UpdateIPAllocation(ctx context.Context, allocation *model.IPAllocation) error

	// DeleteIPAllocation deletes an IP allocation by ID.
	DeleteIPAllocation(ctx context.Context, id string) error

	// DeleteIPAllocationByPeerID deletes an IP allocation by peer ID (hard delete).
	DeleteIPAllocationByPeerID(ctx context.Context, peerID string) error

	// ListIPAllocations lists IP allocations with optional filters and pagination.
	ListIPAllocations(ctx context.Context, opt IPAllocationListOptions) ([]*model.IPAllocation, int64, error)

	// GetAllocatedIPsByPoolID retrieves all allocated IP addresses for a given IP pool.
	GetAllocatedIPsByPoolID(ctx context.Context, poolID string) ([]string, error)
}

// IPAllocationListOptions defines options for listing IP allocations.
type IPAllocationListOptions struct {
	IPPoolID string
	PeerID   string
	Status   string
	Offset   int
	Limit    int
}

