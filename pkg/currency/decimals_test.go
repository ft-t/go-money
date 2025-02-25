package currency_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/currency"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/shopspring/decimal"
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

func TestToString(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		ct := &database.Currency{
			ID:            "USD",
			DecimalPlaces: 3,
		}
		assert.NoError(t, gormDB.Create(ct).Error)

		dec := currency.NewDecimalService()

		res := dec.ToString(context.TODO(), decimal.RequireFromString("200"), "USD")
		assert.EqualValues(t, "200.000", res)

		res = dec.ToString(context.TODO(), decimal.RequireFromString("200"), "USD") // from cache
		assert.EqualValues(t, "200.000", res)
	})
}
