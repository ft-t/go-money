package configuration

import (
	"context"
	"github.com/sethvargo/go-envconfig"
)

var configuration *Configuration

func GetConfiguration() *Configuration {
	if configuration != nil {
		return configuration
	}
	
	var cfg Configuration

	if err := envconfig.ProcessWith(context.Background(), &envconfig.Config{
		Target: &cfg,
	}); err != nil {
		panic(err)
	}

	configuration = &cfg

	return configuration
}
