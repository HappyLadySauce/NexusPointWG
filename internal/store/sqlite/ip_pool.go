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

type ipPools struct {
	db *gorm.DB
}

func newIPPools(ds *datastore) *ipPools {
	return &ipPools{ds.db}
}

func (i *ipPools) CreateIPPool(ctx context.Context, pool *model.IPPool) error {
	err := i.db.WithContext(ctx).Create(pool).Error
	if err != nil {
		if isUniqueConstraintError(err) {
			return errors.WithCode(code.ErrIPPoolAlreadyExists, "IP pool with this CIDR or name already exists")
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (i *ipPools) GetIPPool(ctx context.Context, id string) (*model.IPPool, error) {
	var pool model.IPPool
	err := i.db.WithContext(ctx).Where("id = ?", id).First(&pool).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrIPPoolNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return &pool, nil
}

func (i *ipPools) GetIPPoolByCIDR(ctx context.Context, cidr string) (*model.IPPool, error) {
	var pool model.IPPool
	err := i.db.WithContext(ctx).Where("cidr = ?", cidr).First(&pool).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrIPPoolNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return &pool, nil
}

func (i *ipPools) UpdateIPPool(ctx context.Context, pool *model.IPPool) error {
	err := i.db.WithContext(ctx).Save(pool).Error
	if err != nil {
		if isUniqueConstraintError(err) {
			return errors.WithCode(code.ErrIPPoolAlreadyExists, "IP pool with this CIDR or name already exists")
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (i *ipPools) DeleteIPPool(ctx context.Context, id string) error {
	// Check if pool is in use
	var count int64
	if err := i.db.WithContext(ctx).Model(&model.IPAllocation{}).Where("ip_pool_id = ? AND status = ?", id, model.IPAllocationStatusAllocated).Count(&count).Error; err != nil {
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	if count > 0 {
		return errors.WithCode(code.ErrIPPoolInUse, "IP pool is in use and cannot be deleted")
	}

	err := i.db.WithContext(ctx).Where("id = ?", id).Delete(&model.IPPool{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // Idempotent delete
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (i *ipPools) ListIPPools(ctx context.Context, opt store.IPPoolListOptions) ([]*model.IPPool, int64, error) {
	var (
		pools []*model.IPPool
		total int64
	)

	dbq := i.db.WithContext(ctx).Model(&model.IPPool{})
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

	if err := dbq.Order("created_at DESC").Offset(offset).Limit(limit).Find(&pools).Error; err != nil {
		return nil, 0, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return pools, total, nil
}

// BatchCreateIPPools creates multiple IP pools in a transaction.
// Returns error if any pool creation fails, causing a rollback of all operations.
func (i *ipPools) BatchCreateIPPools(ctx context.Context, pools []*model.IPPool) error {
	return i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, pool := range pools {
			if err := tx.Create(pool).Error; err != nil {
				if isUniqueConstraintError(err) {
					return errors.WithCode(code.ErrIPPoolAlreadyExists, "IP pool with this CIDR or name already exists")
				}
				return errors.WithCode(code.ErrDatabase, "%s", err.Error())
			}
		}
		return nil
	})
}

// BatchUpdateIPPools updates multiple IP pools in a transaction.
// Returns error if any pool update fails, causing a rollback of all operations.
func (i *ipPools) BatchUpdateIPPools(ctx context.Context, pools []*model.IPPool) error {
	return i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, pool := range pools {
			if err := tx.Save(pool).Error; err != nil {
				if isUniqueConstraintError(err) {
					return errors.WithCode(code.ErrIPPoolAlreadyExists, "IP pool with this CIDR or name already exists")
				}
				return errors.WithCode(code.ErrDatabase, "%s", err.Error())
			}
		}
		return nil
	})
}

// BatchDeleteIPPools deletes multiple IP pools by IDs in a transaction.
// Returns error if any pool deletion fails, causing a rollback of all operations.
func (i *ipPools) BatchDeleteIPPools(ctx context.Context, ids []string) error {
	return i.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			// Check if pool is in use
			var count int64
			if err := tx.Model(&model.IPAllocation{}).Where("ip_pool_id = ? AND status = ?", id, model.IPAllocationStatusAllocated).Count(&count).Error; err != nil {
				return errors.WithCode(code.ErrDatabase, "%s", err.Error())
			}
			if count > 0 {
				return errors.WithCode(code.ErrIPPoolInUse, "IP pool is in use and cannot be deleted")
			}

			if err := tx.Where("id = ?", id).Delete(&model.IPPool{}).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// Continue if record not found (idempotent delete)
					continue
				}
				return errors.WithCode(code.ErrDatabase, "%s", err.Error())
			}
		}
		return nil
	})
}
