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

type wgPeers struct {
	db *gorm.DB
}

func newWGPeers(ds *datastore) *wgPeers {
	return &wgPeers{ds.db}
}

func (w *wgPeers) CreatePeer(ctx context.Context, peer *model.WGPeer) error {
	err := w.db.WithContext(ctx).Create(peer).Error
	if err != nil {
		if isUniqueConstraintError(err) {
			return errors.WithCode(code.ErrIPAlreadyInUse, "peer with this public key already exists")
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (w *wgPeers) GetPeer(ctx context.Context, id string) (*model.WGPeer, error) {
	var peer model.WGPeer
	err := w.db.WithContext(ctx).Where("id = ?", id).First(&peer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrWGPeerNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return &peer, nil
}

func (w *wgPeers) GetPeerByPublicKey(ctx context.Context, publicKey string) (*model.WGPeer, error) {
	var peer model.WGPeer
	err := w.db.WithContext(ctx).Where("client_public_key = ?", publicKey).First(&peer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrWGPeerNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return &peer, nil
}

func (w *wgPeers) UpdatePeer(ctx context.Context, peer *model.WGPeer) error {
	err := w.db.WithContext(ctx).Save(peer).Error
	if err != nil {
		if isUniqueConstraintError(err) {
			return errors.WithCode(code.ErrIPAlreadyInUse, "peer with this public key already exists")
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (w *wgPeers) DeletePeer(ctx context.Context, id string) error {
	err := w.db.WithContext(ctx).Where("id = ?", id).Delete(&model.WGPeer{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // Idempotent delete
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (w *wgPeers) ListPeers(ctx context.Context, opt store.WGPeerListOptions) ([]*model.WGPeer, int64, error) {
	var (
		peers []*model.WGPeer
		total int64
	)

	dbq := w.db.WithContext(ctx).Model(&model.WGPeer{})
	if strings.TrimSpace(opt.UserID) != "" {
		dbq = dbq.Where("user_id = ?", opt.UserID)
	}
	if strings.TrimSpace(opt.Status) != "" {
		dbq = dbq.Where("status = ?", opt.Status)
	}
	if strings.TrimSpace(opt.IPPoolID) != "" {
		dbq = dbq.Where("ip_pool_id = ?", opt.IPPoolID)
	}
	if strings.TrimSpace(opt.DeviceName) != "" {
		dbq = dbq.Where("device_name LIKE ?", "%"+opt.DeviceName+"%")
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

	if err := dbq.Order("created_at DESC").Offset(offset).Limit(limit).Find(&peers).Error; err != nil {
		return nil, 0, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return peers, total, nil
}

func (w *wgPeers) CountPeersByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := w.db.WithContext(ctx).Model(&model.WGPeer{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return 0, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return count, nil
}

// BatchCreatePeers creates multiple WireGuard peers in a transaction.
// Returns error if any peer creation fails, causing a rollback of all operations.
func (w *wgPeers) BatchCreatePeers(ctx context.Context, peers []*model.WGPeer) error {
	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, peer := range peers {
			if err := tx.Create(peer).Error; err != nil {
				if isUniqueConstraintError(err) {
					return errors.WithCode(code.ErrIPAlreadyInUse, "peer with this public key already exists")
				}
				return errors.WithCode(code.ErrDatabase, "%s", err.Error())
			}
		}
		return nil
	})
}

// BatchUpdatePeers updates multiple WireGuard peers in a transaction.
// Returns error if any peer update fails, causing a rollback of all operations.
func (w *wgPeers) BatchUpdatePeers(ctx context.Context, peers []*model.WGPeer) error {
	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, peer := range peers {
			if err := tx.Save(peer).Error; err != nil {
				if isUniqueConstraintError(err) {
					return errors.WithCode(code.ErrIPAlreadyInUse, "peer with this public key already exists")
				}
				return errors.WithCode(code.ErrDatabase, "%s", err.Error())
			}
		}
		return nil
	})
}

// BatchDeletePeers deletes multiple WireGuard peers by IDs in a transaction.
// Returns error if any peer deletion fails, causing a rollback of all operations.
func (w *wgPeers) BatchDeletePeers(ctx context.Context, ids []string) error {
	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			if err := tx.Where("id = ?", id).Delete(&model.WGPeer{}).Error; err != nil {
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
