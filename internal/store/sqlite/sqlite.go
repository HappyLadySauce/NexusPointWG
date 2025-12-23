package sqlite

import (
	"fmt"
	"sync"

	"gorm.io/gorm"

	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/db"
	"github.com/HappyLadySauce/NexusPointWG/internal/store"
	"github.com/HappyLadySauce/NexusPointWG/pkg/options"
	"github.com/HappyLadySauce/errors"
)

type datastore struct {
	db *gorm.DB
}

func (ds *datastore) Users() store.UserStore {
	return newUsers(ds)
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
	if opts == nil && sqliteFactory == nil {
		return nil, errors.New("sqlite options are not set")
	}

	var err error
	var dbIns *gorm.DB
	once.Do(func() {
		options := &db.Options{
			DataSourceName: opts.DataSourceName,
		}
		dbIns, err = db.New(options)
		if err != nil {
			err = errors.Wrap(err, "failed to create sqlite db")
			return
		}
		sqliteFactory = &datastore{dbIns}
	})

	if sqliteFactory == nil || err != nil {
		return nil, fmt.Errorf("failed to get sqlite factory, sqliteFactory: %v, err: %w", sqliteFactory, err)
	}

	return sqliteFactory, nil	
}