package database_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
)

var gormDB *gorm.DB
var cfg *configuration.Configuration

func TestMain(m *testing.M) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestSystemConfig_TableName(t *testing.T) {
	cfg := database.SystemConfig{}
	assert.Equal(t, "system_configurations", cfg.TableName())
}

func TestSystemConfigRepository_Get_Success(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))

	type tc struct {
		name     string
		key      string
		value    string
		expected string
	}

	cases := []tc{
		{
			name:     "existing key",
			key:      "test_key_1",
			value:    "test_value_1",
			expected: "test_value_1",
		},
		{
			name:     "key with special characters",
			key:      "jwt_private_key",
			value:    "-----BEGIN RSA PRIVATE KEY-----\nMIIE...\n-----END RSA PRIVATE KEY-----",
			expected: "-----BEGIN RSA PRIVATE KEY-----\nMIIE...\n-----END RSA PRIVATE KEY-----",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := database.NewSystemConfigRepository(gormDB)

			require.NoError(t, repo.Set(context.Background(), c.key, c.value))

			result, err := repo.Get(context.Background(), c.key)

			require.NoError(t, err)
			assert.Equal(t, c.expected, result)
		})
	}
}

func TestSystemConfigRepository_Get_NotFound(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))

	repo := database.NewSystemConfigRepository(gormDB)

	result, err := repo.Get(context.Background(), "non_existent_key")

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestSystemConfigRepository_Set_Success(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))

	type tc struct {
		name  string
		key   string
		value string
	}

	cases := []tc{
		{
			name:  "new key",
			key:   "new_config_key",
			value: "new_config_value",
		},
		{
			name:  "empty value",
			key:   "empty_value_key",
			value: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := database.NewSystemConfigRepository(gormDB)

			err := repo.Set(context.Background(), c.key, c.value)

			require.NoError(t, err)

			result, getErr := repo.Get(context.Background(), c.key)
			require.NoError(t, getErr)
			assert.Equal(t, c.value, result)
		})
	}
}

func TestSystemConfigRepository_Set_Update(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))

	repo := database.NewSystemConfigRepository(gormDB)
	key := "update_test_key"

	require.NoError(t, repo.Set(context.Background(), key, "initial_value"))

	result1, err := repo.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, "initial_value", result1)

	require.NoError(t, repo.Set(context.Background(), key, "updated_value"))

	result2, err := repo.Get(context.Background(), key)
	require.NoError(t, err)
	assert.Equal(t, "updated_value", result2)
}
