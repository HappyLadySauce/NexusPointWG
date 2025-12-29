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
	return &wgPeers{db: ds.db}
}

func (s *wgPeers) Create(ctx context.Context, peer *model.WGPeer) error {
	if peer == nil {
		return errors.WithCode(code.ErrBind, "peer is nil")
	}
	if err := s.db.WithContext(ctx).Create(peer).Error; err != nil {
		if isUniqueConstraintError(err) {
			// Unify as already exists for now; controller/service can refine mapping later if needed.
			return errors.WithCode(code.ErrDatabase, "%s", err.Error())
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (s *wgPeers) Get(ctx context.Context, id string) (*model.WGPeer, error) {
	var peer model.WGPeer
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&peer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrWGPeerNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return &peer, nil
}

func (s *wgPeers) Update(ctx context.Context, peer *model.WGPeer) error {
	if peer == nil {
		return errors.WithCode(code.ErrBind, "peer is nil")
	}
	if err := s.db.WithContext(ctx).Save(peer).Error; err != nil {
		if isUniqueConstraintError(err) {
			return errors.WithCode(code.ErrDatabase, "%s", err.Error())
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (s *wgPeers) Delete(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.WithCode(code.ErrBind, "id is required")
	}
	if err := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.WGPeer{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (s *wgPeers) List(ctx context.Context, opt store.WGPeerListOptions) ([]*model.WGPeer, int64, error) {
	var (
		peers []*model.WGPeer
		total int64
	)

	dbq := s.db.WithContext(ctx).Model(&model.WGPeer{})
	if strings.TrimSpace(opt.UserID) != "" {
		dbq = dbq.Where("user_id = ?", opt.UserID)
	}
	if strings.TrimSpace(opt.DeviceName) != "" {
		dbq = dbq.Where("device_name LIKE ?", "%"+opt.DeviceName+"%")
	}
	if strings.TrimSpace(opt.ClientIP) != "" {
		dbq = dbq.Where("client_ip = ?", opt.ClientIP)
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
	if limit > 10000 {
		limit = 10000
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
