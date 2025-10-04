package importers_test

import (
	"testing"

	v1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/stretchr/testify/assert"
)

func TestGetAccountMapByNumbers(t *testing.T) {
	bp := importers.NewBaseParser(nil, nil, nil)

	t.Run("single account with single number", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "1234",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, int32(1), result["1234"].ID)
	})

	t.Run("single account with multiple numbers", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "1234,5678",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int32(1), result["1234"].ID)
		assert.Equal(t, int32(1), result["5678"].ID)
	})

	t.Run("multiple accounts with different numbers", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "1234",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
			{
				ID:            2,
				AccountNumber: "5678",
				Currency:      "USD",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int32(1), result["1234"].ID)
		assert.Equal(t, int32(2), result["5678"].ID)
	})

	t.Run("account with empty number gets uuid", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Len(t, result, 1)

		var foundAccount *database.Account
		for _, acc := range result {
			foundAccount = acc
			break
		}
		assert.NotNil(t, foundAccount)
		assert.Equal(t, int32(1), foundAccount.ID)
	})

	t.Run("account with whitespace-only number gets uuid", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "  ",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
	})

	t.Run("duplicate account numbers returns error", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "1234",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
			{
				ID:            2,
				AccountNumber: "1234",
				Currency:      "USD",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "duplicate account number: 1234")
	})

	t.Run("account with comma-separated numbers and whitespace", func(t *testing.T) {
		accounts := []*database.Account{
			{
				ID:            1,
				AccountNumber: "1234 , 5678 , 9012",
				Currency:      "UAH",
				Type:          v1.AccountType_ACCOUNT_TYPE_ASSET,
			},
		}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, int32(1), result["1234"].ID)
		assert.Equal(t, int32(1), result["5678"].ID)
		assert.Equal(t, int32(1), result["9012"].ID)
	})

	t.Run("empty accounts list", func(t *testing.T) {
		accounts := []*database.Account{}

		result, err := bp.GetAccountMapByNumbers(accounts)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}
