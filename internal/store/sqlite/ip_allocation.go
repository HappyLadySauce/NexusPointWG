package sqlite

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/errors"
)

type ipAllocations struct {
	db *gorm.DB
}

func newIPAllocations(ds *datastore) *ipAllocations {
	return &ipAllocations{ds.db}
}

func (i *ipAllocations) CreateIPAllocation(ctx context.Context, allocation *model.IPAllocation) error {
	err := i.db.WithContext(ctx).Create(allocation).Error
	if err != nil {
		if isUniqueConstraintError(err) {
			return errors.WithCode(code.ErrIPAlreadyInUse, "IP address is already allocated")
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (i *ipAllocations) GetIPAllocation(ctx context.Context, id string) (*model.IPAllocation, error) {
	var allocation model.IPAllocation
	err := i.db.WithContext(ctx).Where("id = ?", id).First(&allocation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrDatabase, "IP allocation not found")
		}
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return &allocation, nil
}

func (i *ipAllocations) GetIPAllocationByIPAddress(ctx context.Context, ipAddress string) (*model.IPAllocation, error) {
	var allocation model.IPAllocation
	err := i.db.WithContext(ctx).Where("ip_address = ? AND status = ?", ipAddress, model.IPAllocationStatusAllocated).First(&allocation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found, not an error
		}
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return &allocation, nil
}

func (i *ipAllocations) GetIPAllocationByPeerID(ctx context.Context, peerID string) (*model.IPAllocation, error) {
	var allocation model.IPAllocation
	err := i.db.WithContext(ctx).Where("peer_id = ? AND status = ?", peerID, model.IPAllocationStatusAllocated).First(&allocation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found, not an error
		}
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return &allocation, nil
}

func (i *ipAllocations) UpdateIPAllocation(ctx context.Context, allocation *model.IPAllocation) error {
	err := i.db.WithContext(ctx).Save(allocation).Error
	if err != nil {
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (i *ipAllocations) DeleteIPAllocation(ctx context.Context, id string) error {
	err := i.db.WithContext(ctx).Where("id = ?", id).Delete(&model.IPAllocation{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // Idempotent delete
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (i *ipAllocations) ListIPAllocations(ctx context.Context, opt store.IPAllocationListOptions) ([]*model.IPAllocation, int64, error) {
	var (
		allocations []*model.IPAllocation
		total       int64
	)

	dbq := i.db.WithContext(ctx).Model(&model.IPAllocation{})
	if strings.TrimSpace(opt.IPPoolID) != "" {
		dbq = dbq.Where("ip_pool_id = ?", opt.IPPoolID)
	}
	if strings.TrimSpace(opt.PeerID) != "" {
		dbq = dbq.Where("peer_id = ?", opt.PeerID)
	}
	if strings.TrimSpace(opt.Status) != "" {
		dbq = dbq.Where("status = ?", opt.Status)
	}

	if err := dbq.Count(&total).Error; err != nil {
		return nil, 0, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}

	limit := opt.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	offset := opt.Offset
	if offset < 0 {
		offset = 0
	}

	if err := dbq.Order("created_at DESC").Offset(offset).Limit(limit).Find(&allocations).Error; err != nil {
		return nil, 0, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return allocations, total, nil
}

func (i *ipAllocations) GetAllocatedIPsByPoolID(ctx context.Context, poolID string) ([]string, error) {
	var allocations []model.IPAllocation
	err := i.db.WithContext(ctx).
		Where("ip_pool_id = ? AND status = ?", poolID, model.IPAllocationStatusAllocated).
		Select("ip_address").
		Find(&allocations).Error
	if err != nil {
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}

	ips := make([]string, 0, len(allocations))
	for _, allocation := range allocations {
		ips = append(ips, allocation.IPAddress)
	}
	return ips, nil
}
