package db

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Options struct {
	DataSourceName string
}

// New creates a new GORM database connection using pure Go SQLite driver (no CGO required).
func New(opts *Options) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Dialector{
		DSN: opts.DataSourceName,
	}, &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
