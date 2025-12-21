package database_test

import (
	"testing"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestAppConfig_TableName(t *testing.T) {
	cfg := database.AppConfig{}
	assert.Equal(t, "app_configs", cfg.TableName())
}
