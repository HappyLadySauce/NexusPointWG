package config

import (
	"sync"

	"github.com/HappyLadySauce/NexusPointWG/pkg/options"
)

// Config is the running configuration structure of NexusPointWG.
// It is created from cmd/app/options.Options and stored globally via Init/Get.
type Config struct {
	InsecureServing *options.InsecureServingOptions
	Sqlite          *options.SqliteOptions
	Log             *options.LogOptions
	JWT             *options.JWTOptions
}

var (
	once sync.Once
	cfg  *Config
)

// Init sets the global config. It can be called only once.
func Init(c *Config) {
	once.Do(func() {
		cfg = c
	})
}

// Get returns the global config. It panics if Init() was never called.
func Get() *Config {
	if cfg == nil {
		panic("config is not initialized: call config.Init() before use")
	}
	return cfg
}
