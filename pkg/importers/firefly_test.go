package importers_test

import (
	accountsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/accounts/v1"
	"context"
	"encoding/json"
	"github.com/ft-t/go-money/pkg/accounts"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/currency"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/ft-t/go-money/pkg/mappers"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"net/http"
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

func TestFirefly(t *testing.T) {
	t.Skip("todo")

	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
	data, err := os.ReadFile("C:\\Users\\iqpir\\Downloads\\2025_06_26_transaction_export.csv")
	assert.NoError(t, err)

	accountsData, err := os.ReadFile("C:\\Users\\iqpir\\Result_17.json")
	assert.NoError(t, err)

	var bulkAccounts []*accountsv1.CreateAccountRequest
	assert.NoError(t, json.Unmarshal(accountsData, &bulkAccounts))

	m := mappers.NewMapper(&mappers.MapperConfig{
		DecimalSvc: currency.NewDecimalService(),
	})

	accountSvc := accounts.NewService(&accounts.ServiceConfig{
		MapperSvc: m,
	})
	_, err = accountSvc.CreateBulk(context.TODO(), &accountsv1.CreateAccountsBulkRequest{
		Accounts: bulkAccounts,
	})
	assert.NoError(t, err)

	var allAccounts []*database.Account
	assert.NoError(t, gormDB.Find(&allAccounts).Error)

	cur := currency.NewSyncer(http.DefaultClient, transactions.NewBaseAmountService(), configuration.CurrencyConfig{})
	assert.NoError(t, cur.Sync(context.TODO(), "http://go-money-exchange-rates.s3-website.eu-north-1.amazonaws.com/latest.json"))

	converter := currency.NewConverter()

	txSvc := transactions.NewService(&transactions.ServiceConfig{
		StatsSvc:             transactions.NewStatService(),
		MapperSvc:            m,
		CurrencyConverterSvc: converter,
	})
	importer := importers.NewFireflyImporter(txSvc)

	result, err := importer.Import(context.TODO(), &importers.ImportRequest{
		Data:     data,
		Accounts: allAccounts,
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
}
