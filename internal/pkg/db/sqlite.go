package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Options struct {
	DataSourceName string
}

func New(opts *Options) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(opts.DataSourceName), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}