package options

import (
	"encoding/json"

	"github.com/spf13/pflag"
	"k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"

	"github.com/HappyLadySauce/NexusPointWG/pkg/options"
)

type Options struct {
	InsecureServing *options.InsecureServingOptions `mapstructure:"insecure"`
	Sqlite          *options.SqliteOptions          `mapstructure:"sqlite"`
	Log             *options.LogOptions             `mapstructure:"logs"`
}

func NewOptions() *Options {
	return &Options{
		InsecureServing: options.NewInsecureServingOptions(),
		Sqlite:          options.NewSqliteOptions(),
		Log:             options.NewLogOptions(),
	}
}

// AddFlags adds the flags to the specified FlagSet and returns the grouped flag sets.
func (o *Options) AddFlags(fs *pflag.FlagSet) *flag.NamedFlagSets {
	nfs := &flag.NamedFlagSets{}

	// add the flags to the NamedFlagSets
	configFS := nfs.FlagSet("Config")
	options.AddConfigFlag(configFS)

	insecureServingFS := nfs.FlagSet("Insecure Serving")
	o.InsecureServing.AddFlags(insecureServingFS)

	// Add log file rotation flags
	logsFlagSet := nfs.FlagSet("Logs")
	logs.AddFlags(logsFlagSet)
	o.Log.AddFlags(logsFlagSet)

	// add the flags to the NamedFlagSets
	sqliteFS := nfs.FlagSet("SQLite")
	o.Sqlite.AddFlags(sqliteFS)

	// add the flags to the main Command
	for _, name := range nfs.Order {
		fs.AddFlagSet(nfs.FlagSets[name])
	}
	return nfs
}

func (o *Options) String() string {
	data, _ := json.Marshal(o)

	return string(data)
}
