package configuration

import (
	"context"
	"fmt"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/sethvargo/go-envconfig"
	"os"
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

	if boilerplate.GetCurrentEnvironment() == boilerplate.Ci { // for CI, we use a unique database name to avoid conflicts
		configuration.Db.Db = fmt.Sprintf("ci_%v", int64(os.Getpid()))
		configuration.ReadOnlyDb.Db = configuration.Db.Db
	}

	configuration = &cfg

	return configuration
}
