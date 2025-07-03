package importers_test

import (
	accountsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/accounts/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	v1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	_ "embed"
	"encoding/json"
	"github.com/ft-t/go-money/pkg/accounts"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/currency"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/ft-t/go-money/pkg/mappers"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/golang/mock/gomock"
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

//go:embed testdata/accounts.json
var accountsByteData []byte

//go:embed testdata/ff_withdrawal.csv
var ffWithdrawalByteData []byte

//go:embed testdata/rates.json
var ratesByteData []byte

func TestFireflyImport(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	var rates []*database.Currency
	assert.NoError(t, json.Unmarshal(ratesByteData, &rates))

	assert.NoError(t, gormDB.Create(&rates).Error)

	var accountsData []*database.Account
	assert.NoError(t, json.Unmarshal(accountsByteData, &accountsData))

	assert.NoError(t, gormDB.Create(&accountsData).Error)

	t.Run("basic multi currency withdrawal", func(t *testing.T) {
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		importer := importers.NewFireflyImporter(txSvc)

		txSvc.EXPECT().CreateBulkInternal(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, requests []*transactions.BulkRequest, db *gorm.DB) ([]*transactionsv1.CreateTransactionResponse, error) {
				assert.Len(t, requests, 1)

				tx := requests[0].Req.Transaction.(*transactionsv1.CreateTransactionRequest_Withdrawal)

				assert.EqualValues(t, tx.Withdrawal.SourceCurrency, "UAH")
				assert.EqualValues(t, *tx.Withdrawal.ForeignCurrency, "PLN")

				assert.EqualValues(t, "-964.44", tx.Withdrawal.SourceAmount)
				assert.EqualValues(t, "-83.81", *tx.Withdrawal.ForeignAmount)

				assert.EqualValues(t, accountsData[0].ID, tx.Withdrawal.SourceAccountId)

				assert.EqualValues(t, "firefly_2805", *requests[0].Req.InternalReferenceNumber)

				assert.EqualValues(t, []int32{1}, requests[0].Req.TagIds)

				return []*transactionsv1.CreateTransactionResponse{
					{
						Transaction: &v1.Transaction{
							InternalReferenceNumber: requests[0].Req.InternalReferenceNumber,
						},
					},
				}, nil
			})

		result, err := importer.Import(context.TODO(), &importers.ImportRequest{
			Data:     ffWithdrawalByteData,
			Accounts: accountsData,
			Tags: map[string]*database.Tag{
				"Grocery": {
					ID: 1,
				},
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestFireflyIntegration(t *testing.T) {
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
