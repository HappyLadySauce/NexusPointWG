package service

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
)

// IPPoolSrv defines the interface for IP pool business logic.
type IPPoolSrv interface {
	CreateIPPool(ctx context.Context, pool *model.IPPool) error
	GetIPPool(ctx context.Context, id string) (*model.IPPool, error)
	GetIPPoolByCIDR(ctx context.Context, cidr string) (*model.IPPool, error)
	UpdateIPPool(ctx context.Context, pool *model.IPPool) error
	DeleteIPPool(ctx context.Context, id string) error
	ListIPPools(ctx context.Context, opt store.IPPoolListOptions) ([]*model.IPPool, int64, error)
}

type ipPoolSrv struct {
	store store.Factory
}

// IPPoolSrv if implemented, then ipPoolSrv implements IPPoolSrv interface.
var _ IPPoolSrv = (*ipPoolSrv)(nil)

func newIPPools(s *service) *ipPoolSrv {
	return &ipPoolSrv{store: s.store}
}

func (i *ipPoolSrv) CreateIPPool(ctx context.Context, pool *model.IPPool) error {
	return i.store.IPPools().CreateIPPool(ctx, pool)
}

func (i *ipPoolSrv) GetIPPool(ctx context.Context, id string) (*model.IPPool, error) {
	return i.store.IPPools().GetIPPool(ctx, id)
}

func (i *ipPoolSrv) GetIPPoolByCIDR(ctx context.Context, cidr string) (*model.IPPool, error) {
	return i.store.IPPools().GetIPPoolByCIDR(ctx, cidr)
}

func (i *ipPoolSrv) UpdateIPPool(ctx context.Context, pool *model.IPPool) error {
	return i.store.IPPools().UpdateIPPool(ctx, pool)
}

func (i *ipPoolSrv) DeleteIPPool(ctx context.Context, id string) error {
	return i.store.IPPools().DeleteIPPool(ctx, id)
}

func (i *ipPoolSrv) ListIPPools(ctx context.Context, opt store.IPPoolListOptions) ([]*model.IPPool, int64, error) {
	return i.store.IPPools().ListIPPools(ctx, opt)
}
