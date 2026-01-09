package store

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
)

// IPPoolStore defines the interface for IP pool data access.
type IPPoolStore interface {
	// CreateIPPool creates a new IP pool.
	CreateIPPool(ctx context.Context, pool *model.IPPool) error

	// GetIPPool retrieves an IP pool by ID.
	GetIPPool(ctx context.Context, id string) (*model.IPPool, error)

	// GetIPPoolByCIDR retrieves an IP pool by CIDR.
	GetIPPoolByCIDR(ctx context.Context, cidr string) (*model.IPPool, error)

	// UpdateIPPool updates an existing IP pool.
	UpdateIPPool(ctx context.Context, pool *model.IPPool) error

	// DeleteIPPool deletes an IP pool by ID.
	DeleteIPPool(ctx context.Context, id string) error

	// ListIPPools lists IP pools with optional filters and pagination.
	ListIPPools(ctx context.Context, opt IPPoolListOptions) ([]*model.IPPool, int64, error)
	// BatchCreateIPPools creates multiple IP pools in a transaction. Returns error if any pool creation fails.
	BatchCreateIPPools(ctx context.Context, pools []*model.IPPool) error
	// BatchUpdateIPPools updates multiple IP pools in a transaction. Returns error if any pool update fails.
	BatchUpdateIPPools(ctx context.Context, pools []*model.IPPool) error
	// BatchDeleteIPPools deletes multiple IP pools by IDs in a transaction. Returns error if any pool deletion fails.
	BatchDeleteIPPools(ctx context.Context, ids []string) error
}

// IPPoolListOptions defines options for listing IP pools.
type IPPoolListOptions struct {
	Status string
	Offset int
	Limit  int
}
