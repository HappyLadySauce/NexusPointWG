package config

import "github.com/HappyLadySauce/NexusPointWG/cmd/app/options"

type Config struct {
	*options.Options
}

func CreateConfigFromOptions(opts *options.Options) (*Config, error) {
	return &Config{Options: opts}, nil
}