package database_test

import (
	"testing"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestIsDefault(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		account := &database.Account{
			ID:       1,
			Name:     "Default Account",
			Currency: "USD",
			Flags:    database.AccountFlagIsDefault | 2,
		}

		assert.True(t, account.IsDefault())
	})
	t.Run("false", func(t *testing.T) {
		account := &database.Account{
			ID:       2,
			Name:     "Non-default Account",
			Currency: "USD",
			Flags:    2,
		}

		assert.False(t, account.IsDefault())
	})
}
