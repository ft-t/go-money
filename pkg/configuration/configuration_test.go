package configuration_test

import (
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Setenv("ENVIRONMENT", "ci")

		cfg := configuration.GetConfiguration()
		assert.Contains(t, cfg.Db.Db, "ci_")
		assert.Contains(t, cfg.ReadOnlyDb.Db, "ci_")

		cfg2 := configuration.GetConfiguration() // from var
		assert.Equal(t, cfg, cfg2)
	})
}
