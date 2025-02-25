package currency_test

import (
	"context"
	currencyv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/currency/v1"
	v1 "github.com/ft-t/go-money-pb/gen/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/currency"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestGetCurrencies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		cur1 := &database.Currency{
			ID:            "USD",
			Rate:          decimal.NewFromInt(2),
			IsActive:      true,
			DecimalPlaces: 4,
		}
		cur2 := &database.Currency{
			ID:       "EUR",
			Rate:     decimal.NewFromInt(3),
			IsActive: true,
		}
		cur3 := &database.Currency{
			ID:   "GBP",
			Rate: decimal.NewFromInt(4),
			DeletedAt: gorm.DeletedAt{
				Time:  time.Now(),
				Valid: true,
			},
		}

		assert.NoError(t, gormDB.Create(cur1).Error)
		assert.NoError(t, gormDB.Create(cur2).Error)
		assert.NoError(t, gormDB.Create(cur3).Error)

		srv := currency.NewService()

		resp, err := srv.GetCurrencies(context.TODO(), nil)
		assert.NoError(t, err)

		assert.Len(t, resp.Currencies, 3)

		assert.Equal(t, "USD", resp.Currencies[0].Id)
		assert.EqualValues(t, "2.0000", resp.Currencies[0].Rate)
		assert.True(t, resp.Currencies[0].IsActive)
		assert.EqualValues(t, 4, resp.Currencies[0].DecimalPlaces)

		assert.Equal(t, "EUR", resp.Currencies[1].Id)
		assert.Equal(t, "GBP", resp.Currencies[2].Id)
	})
}

func TestCreateCurrency(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		srv := currency.NewService()
		resp, err := srv.CreateCurrency(context.TODO(), &currencyv1.CreateCurrencyRequest{
			Currency: &v1.Currency{
				Id:            "USD",
				Rate:          "5.21",
				IsActive:      true,
				DecimalPlaces: 2,
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		var cur database.Currency
		assert.NoError(t, gormDB.Where("id = ?", "USD").First(&cur).Error)

		assert.Equal(t, "USD", cur.ID)
		assert.EqualValues(t, "5.21", cur.Rate.String())
		assert.True(t, cur.IsActive)
		assert.EqualValues(t, 2, cur.DecimalPlaces)
	})

	t.Run("fail duplicate", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		srv := currency.NewService()
		resp, err := srv.CreateCurrency(context.TODO(), &currencyv1.CreateCurrencyRequest{
			Currency: &v1.Currency{
				Id:            "USD",
				Rate:          "5.21",
				IsActive:      true,
				DecimalPlaces: 2,
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		resp, err = srv.CreateCurrency(context.TODO(), &currencyv1.CreateCurrencyRequest{
			Currency: &v1.Currency{
				Id:            "USD",
				Rate:          "5.21",
				IsActive:      true,
				DecimalPlaces: 2,
			},
		})

		assert.ErrorContains(t, err, "duplicate key value violates unique constraint")
		assert.Nil(t, resp)
	})

	t.Run("fail rate", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		srv := currency.NewService()
		resp, err := srv.CreateCurrency(context.TODO(), &currencyv1.CreateCurrencyRequest{
			Currency: &v1.Currency{
				Id:            "USD",
				Rate:          "x5.21",
				IsActive:      true,
				DecimalPlaces: 2,
			},
		})

		assert.ErrorContains(t, err, "can't convert x5.21 to decimal")
		assert.Nil(t, resp)
	})
}

func TestUpdateCurrency(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		cur := &database.Currency{
			ID:            "USD",
			Rate:          decimal.NewFromInt(2),
			IsActive:      true,
			DecimalPlaces: 4,
		}

		assert.NoError(t, gormDB.Create(cur).Error)

		srv := currency.NewService()
		resp, err := srv.UpdateCurrency(context.TODO(), &currencyv1.UpdateCurrencyRequest{
			Currency: &v1.Currency{
				Id:            "USD",
				Rate:          "5.21",
				IsActive:      false,
				DecimalPlaces: 2,
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		var cur2 database.Currency
		assert.NoError(t, gormDB.Where("id = ?", "USD").First(&cur2).Error)

		assert.Equal(t, "USD", cur2.ID)
		assert.EqualValues(t, "5.21", cur2.Rate.String())
		assert.False(t, cur2.IsActive)
		assert.EqualValues(t, 2, cur2.DecimalPlaces)
	})

	t.Run("fail not found", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		srv := currency.NewService()
		resp, err := srv.UpdateCurrency(context.TODO(), &currencyv1.UpdateCurrencyRequest{
			Currency: &v1.Currency{
				Id:            "USD",
				Rate:          "5.21",
				IsActive:      false,
				DecimalPlaces: 2,
			},
		})

		assert.ErrorContains(t, err, "record not found")
		assert.Nil(t, resp)
	})

	t.Run("fail rate", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		cur := &database.Currency{
			ID:            "USD",
			Rate:          decimal.NewFromInt(2),
			IsActive:      true,
			DecimalPlaces: 4,
		}

		assert.NoError(t, gormDB.Create(cur).Error)

		srv := currency.NewService()
		resp, err := srv.UpdateCurrency(context.TODO(), &currencyv1.UpdateCurrencyRequest{
			Currency: &v1.Currency{
				Id:            "USD",
				Rate:          "x5.21",
				IsActive:      false,
				DecimalPlaces: 2,
			},
		})

		assert.ErrorContains(t, err, "can't convert x5.21 to decimal")
		assert.Nil(t, resp)
	})
}
