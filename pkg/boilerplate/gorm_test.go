package boilerplate_test

import (
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
)

var gormDB *gorm.DB
var cfg *configuration.Configuration

func TestMain(m *testing.M) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestGorm(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfgCopy := cfg.Db
		cfgCopy.MaxConnectionLifetimeSec = 10
		cfgCopy.MaxIdleConnections = 100
		cfgCopy.MaxOpenConnections = 1
		cfgCopy.MaxConnectionIdleSec = 100

		resp, err := boilerplate.GetGormConnection(cfgCopy)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("non existing db", func(t *testing.T) {
		cfgCopy := cfg.Db
		cfgCopy.Db = "non_existent_db"

		resp, err := boilerplate.GetGormConnection(cfgCopy)
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "FATAL: database \"non_existent_db\" does not exist")
	})
}
