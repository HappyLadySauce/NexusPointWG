package options

import (
	"fmt"
	"github.com/spf13/pflag"
)

type SqliteOptions struct {
	DataSourceName string `json:"data-source-name" mapstructure:"data-source-name"`
}

func NewSqliteOptions() *SqliteOptions {
	return &SqliteOptions{
		DataSourceName: "./nexuspointwg.db",
	}
}

func (o *SqliteOptions) Validate() []error {
	var errors []error
	if o.DataSourceName == "" {
		errors = append(errors, fmt.Errorf("data-source-name is required"))
	}
	return errors
}

func (o *SqliteOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.DataSourceName, "data-source-name", "", "Data source name for SQLite")
}
