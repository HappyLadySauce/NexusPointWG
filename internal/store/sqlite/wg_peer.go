package sqlite

import (
	"context"

	"gorm.io/gorm"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/errors"
)

type wgPeers struct {
	db *gorm.DB
}

func newWGPeers(ds *datastore) *wgPeers {
	return &wgPeers{ds.db}
}

func (w *wgPeers) CreatePeer(ctx context.Context, peer *model.WGPeer) error {
	if err := w.db.WithContext(ctx).Create(peer).Error; err != nil {
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (w *wgPeers) GetPeer(ctx context.Context, id string) (*model.WGPeer, error) {
	var peer model.WGPeer
	err := w.db.WithContext(ctx).Where("id = ?", id).First(&peer).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrPageNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return &peer, nil
}

func (w *wgPeers) ListPeers(ctx context.Context) ([]*model.WGPeer, error) {
	var peers []*model.WGPeer
	if err := w.db.WithContext(ctx).Order("created_at desc").Find(&peers).Error; err != nil {
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return peers, nil
}

func (w *wgPeers) ListPeersByUser(ctx context.Context, userID string) ([]*model.WGPeer, error) {
	var peers []*model.WGPeer
	if err := w.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&peers).Error; err != nil {
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return peers, nil
}

func (w *wgPeers) UpdatePeer(ctx context.Context, peer *model.WGPeer) error {
	if err := w.db.WithContext(ctx).Save(peer).Error; err != nil {
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}

func (w *wgPeers) DeletePeer(ctx context.Context, id string) error {
	if err := w.db.WithContext(ctx).Where("id = ?", id).Delete(&model.WGPeer{}).Error; err != nil {
		return errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return nil
}
