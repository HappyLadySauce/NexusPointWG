package sqlite

import (
	"sync"

	"gorm.io/gorm"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/db"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/model"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/options"
	"github.com/HappyLadySauce/errors"
	"k8s.io/klog/v2"
)

type datastore struct {
	db *gorm.DB
}

func (ds *datastore) Users() store.UserStore {
	return newUsers(ds)
}

func (ds *datastore) WGPeers() store.WGPeerStore {
	return newWGPeers(ds)
}

func (ds *datastore) Close() error {
	sqlDB, err := ds.db.DB()
	if err != nil {
		return errors.Wrap(err, "failed to get sql db")
	}

	return sqlDB.Close()
}

var (
	sqliteFactory store.Factory
	once          sync.Once
)

func GetSqliteFactoryOr(opts *options.SqliteOptions) (store.Factory, error) {
	// If opts is nil, use default options
	if opts == nil {
		opts = options.NewSqliteOptions()
	}

	var err error
	var dbIns *gorm.DB
	once.Do(func() {
		dbOpts := &db.Options{
			DataSourceName: opts.DataSourceName,
		}
		dbIns, err = db.New(dbOpts)
		if err != nil {
			// Preserve the original error with full context
			err = errors.Wrap(err, "failed to create sqlite db with data source")
			return
		}

		// Auto migrate database schema
		if err = dbIns.AutoMigrate(&model.User{}, &model.WGPeer{}); err != nil {
			err = errors.Wrap(err, "failed to auto migrate database schema")
			return
		}

		klog.V(1).InfoS("Database schema migrated successfully")
		sqliteFactory = &datastore{dbIns}
	})

	if sqliteFactory == nil {
		if err != nil {
			// Return the wrapped error directly to preserve the full error chain
			return nil, errors.Wrap(err, "failed to get sqlite factory")
		}
		// If err is nil but sqliteFactory is nil, create a new error
		return nil, errors.New("failed to get sqlite factory: sqliteFactory is nil but no error was returned")
	}

	return sqliteFactory, nil
}
