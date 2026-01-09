package service

import (
	"context"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/core/ip"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/core/wireguard"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/config"
	"github.com/HappyLadySauce/NexusPointWG/pkg/utils/network"
	"k8s.io/klog/v2"
)

// IPPoolSrv defines the interface for IP pool business logic.
type IPPoolSrv interface {
	CreateIPPool(ctx context.Context, pool *model.IPPool) error
	GetIPPool(ctx context.Context, id string) (*model.IPPool, error)
	GetIPPoolByCIDR(ctx context.Context, cidr string) (*model.IPPool, error)
	UpdateIPPool(ctx context.Context, pool *model.IPPool) error
	DeleteIPPool(ctx context.Context, id string) error
	ListIPPools(ctx context.Context, opt store.IPPoolListOptions) ([]*model.IPPool, int64, error)
	GetAvailableIPs(ctx context.Context, poolID string, limit int) ([]string, error)
	HasAllocatedIPs(ctx context.Context, poolID string) (bool, error)
	UpdateIPPoolsEndpointForGlobalConfigChange(ctx context.Context) error
	UpdateIPPoolsDNSForGlobalConfigChange(ctx context.Context) error
	// BatchCreateIPPools creates multiple IP pools in a transaction.
	BatchCreateIPPools(ctx context.Context, pools []*model.IPPool) error
	// BatchUpdateIPPools updates multiple IP pools in a transaction.
	BatchUpdateIPPools(ctx context.Context, pools []*model.IPPool) error
	// BatchDeleteIPPools deletes multiple IP pools by IDs in a transaction.
	BatchDeleteIPPools(ctx context.Context, ids []string) error
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

// HasAllocatedIPs checks if an IP pool has any allocated IPs.
func (i *ipPoolSrv) HasAllocatedIPs(ctx context.Context, poolID string) (bool, error) {
	allocatedIPs, err := i.store.IPAllocations().GetAllocatedIPsByPoolID(ctx, poolID)
	if err != nil {
		return false, err
	}
	return len(allocatedIPs) > 0, nil
}

func (i *ipPoolSrv) ListIPPools(ctx context.Context, opt store.IPPoolListOptions) ([]*model.IPPool, int64, error) {
	return i.store.IPPools().ListIPPools(ctx, opt)
}

// GetAvailableIPs returns a list of available IP addresses in the pool.
func (i *ipPoolSrv) GetAvailableIPs(ctx context.Context, poolID string, limit int) ([]string, error) {
	allocator := ip.NewAllocator(i.store)
	return allocator.GetAvailableIPs(ctx, poolID, limit)
}

// UpdateIPPoolsEndpointForGlobalConfigChange updates all IP pools that use default endpoint
// when global config (ServerIP, ListenPort, or Endpoint) changes.
func (i *ipPoolSrv) UpdateIPPoolsEndpointForGlobalConfigChange(ctx context.Context) error {
	// Get all IP pools
	pools, _, err := i.store.IPPools().ListIPPools(ctx, store.IPPoolListOptions{})
	if err != nil {
		return err
	}

	// Get global config
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return nil
	}
	wgOpts := cfg.WireGuard

	// Get server IP
	serverIP := wgOpts.ServerIP
	if serverIP == "" {
		detectedIP, err := network.GetServerIP(ctx, "")
		if err == nil {
			serverIP = detectedIP
		}
	}

	// Create config manager for calculating endpoint
	var configManager *wireguard.ServerConfigManager
	if wgOpts.ServerConfigPath() != "" {
		configManager = wireguard.NewServerConfigManager(wgOpts.ServerConfigPath(), wgOpts.ApplyMethod)
	}

	// Update each pool if needed
	for _, pool := range pools {
		needsUpdate := false

		// Check if pool uses default endpoint
		if pool.Endpoint == "" {
			needsUpdate = true
		} else {
			// Check if endpoint matches server IP pattern
			endpointIP, err := ip.ExtractIPFromEndpoint(pool.Endpoint)
			if err == nil && endpointIP != "" {
				// Check if it matches current server IP
				if serverIP != "" && endpointIP == serverIP {
					// Endpoint uses server IP, so it's based on global config
					needsUpdate = true
				} else if pool.Endpoint == wgOpts.Endpoint {
					// Endpoint equals global endpoint config
					needsUpdate = true
				}
			}
		}

		if needsUpdate {
			// Recalculate endpoint
			oldEndpoint := pool.Endpoint
			pool.Endpoint = CalculateIPPoolEndpoint("", wgOpts, configManager, ctx)

			// Only update if endpoint actually changed
			if oldEndpoint != pool.Endpoint {
				// Update in database
				if err := i.store.IPPools().UpdateIPPool(ctx, pool); err != nil {
					klog.V(1).InfoS("failed to update IP pool endpoint", "poolID", pool.ID, "error", err)
					continue
				}

				// If endpoint changed, update all peers using this pool
				// This is handled by the peer update method, but we can also trigger it here
				// For now, the peer update will handle it when it recalculates
			}
		}
	}

	return nil
}

// UpdateIPPoolsDNSForGlobalConfigChange updates all IP pools that use default DNS
// when global config DNS changes.
func (i *ipPoolSrv) UpdateIPPoolsDNSForGlobalConfigChange(ctx context.Context) error {
	// Get all IP pools
	pools, _, err := i.store.IPPools().ListIPPools(ctx, store.IPPoolListOptions{})
	if err != nil {
		return err
	}

	// Get global config
	cfg := config.Get()
	if cfg == nil || cfg.WireGuard == nil {
		return nil
	}
	wgOpts := cfg.WireGuard

	// Update each pool if needed
	for _, pool := range pools {
		// Check if pool uses default DNS (empty means using default)
		if pool.DNS == "" {
			// Recalculate DNS - for IP Pool, if DNS is empty, it should use global DNS
			oldDNS := pool.DNS
			pool.DNS = wgOpts.DNS

			// Only update if DNS actually changed
			if oldDNS != pool.DNS {
				// Update in database
				if err := i.store.IPPools().UpdateIPPool(ctx, pool); err != nil {
					klog.V(1).InfoS("failed to update IP pool DNS", "poolID", pool.ID, "error", err)
					continue
				}

				// If DNS changed, update all peers using this pool
				// This will be handled by UpdatePeersForIPPoolChange
				// We can call it here to ensure peers are updated
				// But to avoid circular dependency, we'll let the peer update handle it
			}
		}
	}

	return nil
}

// BatchCreateIPPools creates multiple IP pools in a transaction.
func (i *ipPoolSrv) BatchCreateIPPools(ctx context.Context, pools []*model.IPPool) error {
	return i.store.IPPools().BatchCreateIPPools(ctx, pools)
}

// BatchUpdateIPPools updates multiple IP pools in a transaction.
func (i *ipPoolSrv) BatchUpdateIPPools(ctx context.Context, pools []*model.IPPool) error {
	return i.store.IPPools().BatchUpdateIPPools(ctx, pools)
}

// BatchDeleteIPPools deletes multiple IP pools by IDs in a transaction.
func (i *ipPoolSrv) BatchDeleteIPPools(ctx context.Context, ids []string) error {
	return i.store.IPPools().BatchDeleteIPPools(ctx, ids)
}
