package importers_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	accountsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/accounts/v1"
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	v1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
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

//go:embed testdata/ff_open_balance_debt.csv
var ffOpenBalanceDebtByteData []byte

//go:embed testdata/ff_open_balance.csv
var ffOpenBalanceByteData []byte

//go:embed testdata/ff_reconciliation.csv
var ffReconciliationByteData []byte

//go:embed testdata/ff_reconciliation_plus.csv
var ffReconciliationPlusByteData []byte

//go:embed testdata/ff_transfer.csv
var ffTransferByteData []byte

//go:embed testdata/ff_withdrawal_debt.csv
var ffWithdrawalDebt []byte

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

				tx := requests[0].Req.Transaction.(*transactionsv1.CreateTransactionRequest_Expense)

				txDate := requests[0].Req.TransactionDate.AsTime().Format(time.RFC3339)
				assert.EqualValues(t, "2025-06-17T13:07:46Z", txDate)

				assert.EqualValues(t, tx.Expense.SourceCurrency, "UAH")
				assert.EqualValues(t, *tx.Expense.FxSourceCurrency, "PLN")

				assert.EqualValues(t, "-964.44", tx.Expense.SourceAmount)
				assert.EqualValues(t, "-83.81", *tx.Expense.FxSourceAmount)

				assert.EqualValues(t, accountsData[0].ID, tx.Expense.SourceAccountId)

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

	t.Run("open balance (debt)", func(t *testing.T) {
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		importer := importers.NewFireflyImporter(txSvc)

		txSvc.EXPECT().CreateBulkInternal(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, requests []*transactions.BulkRequest, db *gorm.DB) ([]*transactionsv1.CreateTransactionResponse, error) {
				assert.Len(t, requests, 1)

				tx := requests[0].Req.Transaction.(*transactionsv1.CreateTransactionRequest_Expense)

				assert.EqualValues(t, tx.Expense.SourceCurrency, "PLN")

				assert.EqualValues(t, "-3900", tx.Expense.SourceAmount)

				assert.EqualValues(t, accountsData[1].ID, tx.Expense.SourceAccountId)

				assert.EqualValues(t, "firefly_1869", *requests[0].Req.InternalReferenceNumber)

				return []*transactionsv1.CreateTransactionResponse{
					{
						Transaction: &v1.Transaction{
							InternalReferenceNumber: requests[0].Req.InternalReferenceNumber,
						},
					},
				}, nil
			})

		result, err := importer.Import(context.TODO(), &importers.ImportRequest{
			Data:     ffOpenBalanceDebtByteData,
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

	t.Run("open balance", func(t *testing.T) {
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		importer := importers.NewFireflyImporter(txSvc)

		txSvc.EXPECT().CreateBulkInternal(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, requests []*transactions.BulkRequest, db *gorm.DB) ([]*transactionsv1.CreateTransactionResponse, error) {
				assert.Len(t, requests, 1)

				tx := requests[0].Req.Transaction.(*transactionsv1.CreateTransactionRequest_Income)

				assert.EqualValues(t, tx.Income.DestinationCurrency, "PLN")

				assert.EqualValues(t, "3520.42", tx.Income.DestinationAmount)

				assert.EqualValues(t, accountsData[1].ID, tx.Income.DestinationAccountId)

				assert.EqualValues(t, "firefly_18691", *requests[0].Req.InternalReferenceNumber)

				return []*transactionsv1.CreateTransactionResponse{
					{
						Transaction: &v1.Transaction{
							InternalReferenceNumber: requests[0].Req.InternalReferenceNumber,
						},
					},
				}, nil
			})

		result, err := importer.Import(context.TODO(), &importers.ImportRequest{
			Data:     ffOpenBalanceByteData,
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

	t.Run("reconciliation (minus)", func(t *testing.T) {
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		importer := importers.NewFireflyImporter(txSvc)

		txSvc.EXPECT().CreateBulkInternal(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, requests []*transactions.BulkRequest, db *gorm.DB) ([]*transactionsv1.CreateTransactionResponse, error) {
				assert.Len(t, requests, 1)

				tx := requests[0].Req.Transaction.(*transactionsv1.CreateTransactionRequest_Adjustment)

				assert.EqualValues(t, "USD", tx.Adjustment.DestinationCurrency)

				assert.EqualValues(t, "-296", tx.Adjustment.DestinationAmount)

				assert.EqualValues(t, accountsData[2].ID, tx.Adjustment.DestinationAccountId)

				assert.EqualValues(t, "firefly_2848", *requests[0].Req.InternalReferenceNumber)

				return []*transactionsv1.CreateTransactionResponse{
					{
						Transaction: &v1.Transaction{
							InternalReferenceNumber: requests[0].Req.InternalReferenceNumber,
						},
					},
				}, nil
			})

		result, err := importer.Import(context.TODO(), &importers.ImportRequest{
			Data:     ffReconciliationByteData,
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

	t.Run("reconciliation (plus)", func(t *testing.T) {
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		importer := importers.NewFireflyImporter(txSvc)

		txSvc.EXPECT().CreateBulkInternal(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, requests []*transactions.BulkRequest, db *gorm.DB) ([]*transactionsv1.CreateTransactionResponse, error) {
				assert.Len(t, requests, 1)

				tx := requests[0].Req.Transaction.(*transactionsv1.CreateTransactionRequest_Adjustment)

				assert.EqualValues(t, "USD", tx.Adjustment.DestinationCurrency)

				assert.EqualValues(t, "49.37", tx.Adjustment.DestinationAmount)

				assert.EqualValues(t, accountsData[2].ID, tx.Adjustment.DestinationAccountId)

				assert.EqualValues(t, "firefly_2830", *requests[0].Req.InternalReferenceNumber)

				return []*transactionsv1.CreateTransactionResponse{
					{
						Transaction: &v1.Transaction{
							InternalReferenceNumber: requests[0].Req.InternalReferenceNumber,
						},
					},
				}, nil
			})

		result, err := importer.Import(context.TODO(), &importers.ImportRequest{
			Data:     ffReconciliationPlusByteData,
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

	t.Run("transfer (same currency)", func(t *testing.T) {
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		importer := importers.NewFireflyImporter(txSvc)

		txSvc.EXPECT().CreateBulkInternal(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, requests []*transactions.BulkRequest, db *gorm.DB) ([]*transactionsv1.CreateTransactionResponse, error) {
				assert.Len(t, requests, 1)

				tx := requests[0].Req.Transaction.(*transactionsv1.CreateTransactionRequest_TransferBetweenAccounts)

				assert.EqualValues(t, accountsData[2].ID, tx.TransferBetweenAccounts.SourceAccountId)
				assert.EqualValues(t, "-1000", tx.TransferBetweenAccounts.SourceAmount)
				assert.EqualValues(t, "USD", tx.TransferBetweenAccounts.SourceCurrency)

				assert.EqualValues(t, "USD", tx.TransferBetweenAccounts.DestinationCurrency)
				assert.EqualValues(t, "1000", tx.TransferBetweenAccounts.DestinationAmount)

				assert.EqualValues(t, accountsData[3].ID, tx.TransferBetweenAccounts.DestinationAccountId)

				assert.EqualValues(t, "firefly_2856", *requests[0].Req.InternalReferenceNumber)

				return []*transactionsv1.CreateTransactionResponse{
					{
						Transaction: &v1.Transaction{
							InternalReferenceNumber: requests[0].Req.InternalReferenceNumber,
						},
					},
				}, nil
			})

		result, err := importer.Import(context.TODO(), &importers.ImportRequest{
			Data:     ffTransferByteData,
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

	t.Run("debt withdrawal -> transfer", func(t *testing.T) {
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		importer := importers.NewFireflyImporter(txSvc)

		txSvc.EXPECT().CreateBulkInternal(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, requests []*transactions.BulkRequest, db *gorm.DB) ([]*transactionsv1.CreateTransactionResponse, error) {
				assert.Len(t, requests, 1)

				tx := requests[0].Req.Transaction.(*transactionsv1.CreateTransactionRequest_TransferBetweenAccounts)

				assert.EqualValues(t, 55, *requests[0].Req.CategoryId)
				assert.EqualValues(t, accountsData[2].ID, tx.TransferBetweenAccounts.SourceAccountId)
				assert.EqualValues(t, "-200", tx.TransferBetweenAccounts.SourceAmount)
				assert.EqualValues(t, "USD", tx.TransferBetweenAccounts.SourceCurrency)

				assert.EqualValues(t, "USD", tx.TransferBetweenAccounts.DestinationCurrency)
				assert.EqualValues(t, "200", tx.TransferBetweenAccounts.DestinationAmount)
				assert.EqualValues(t, accountsData[4].ID, tx.TransferBetweenAccounts.DestinationAccountId)

				assert.EqualValues(t, "firefly_2004", *requests[0].Req.InternalReferenceNumber)

				return []*transactionsv1.CreateTransactionResponse{
					{
						Transaction: &v1.Transaction{
							InternalReferenceNumber: requests[0].Req.InternalReferenceNumber,
						},
					},
				}, nil
			})

		result, err := importer.Import(context.TODO(), &importers.ImportRequest{
			Data:     ffWithdrawalDebt,
			Accounts: accountsData,
			Categories: map[string]*database.Category{
				"Debt repayment": {
					ID:   55,
					Name: "Debt repayment",
				},
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

}

func TestSkipDuplicate(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	var rates []*database.Currency
	assert.NoError(t, json.Unmarshal(ratesByteData, &rates))

	assert.NoError(t, gormDB.Create(&rates).Error)

	var accountsData []*database.Account
	assert.NoError(t, json.Unmarshal(accountsByteData, &accountsData))

	assert.NoError(t, gormDB.Create(&accountsData).Error)

	t.Run("skip duplicate transactions", func(t *testing.T) {
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		importer := importers.NewFireflyImporter(txSvc)

		tx := &database.ImportDeduplication{
			ImportSource: importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
			Key:          "firefly_2805",
		}
		assert.NoError(t, gormDB.Create(&tx).Error)

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
		assert.EqualValues(t, result.DuplicateCount, 1)
	})
}

func TestFireflyNoData(t *testing.T) {
	t.Run("no data", func(t *testing.T) {
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		importer := importers.NewFireflyImporter(txSvc)

		result, err := importer.Import(context.TODO(), &importers.ImportRequest{
			Data:     []byte{},
			Accounts: nil,
			Tags:     nil,
		})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.EqualError(t, err, "no records found in CSV data")
	})
}

func TestFireflyIntegration(t *testing.T) {
	t.Skip("todo")

	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))
	data, err := os.ReadFile("E:\\extra-data\\first.csv")
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

	cur := currency.NewSyncer(http.DefaultClient, transactions.NewBaseAmountService("USD"), configuration.CurrencyConfig{
		BaseCurrency: "USD",
	})
	assert.NoError(t, cur.Sync(context.TODO(), "http://go-money-exchange-rates.s3-website.eu-north-1.amazonaws.com/latest.json"))

	converter := currency.NewConverter("USD")

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

func TestFireflyImport_FailCases(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	var accountsData []*database.Account
	assert.NoError(t, json.Unmarshal(accountsByteData, &accountsData))
	assert.NoError(t, gormDB.Create(&accountsData).Error)

	txSvc := NewMockTransactionSvc(gomock.NewController(t))
	importer := importers.NewFireflyImporter(txSvc)

	t.Run("missing source account", func(t *testing.T) {
		// Withdrawal with unknown account name
		csv := `user_id,group_id,journal_id,created_at,updated_at,group_title,type,amount,foreign_amount,currency_code,foreign_currency_code,description,date,source_name,source_iban,source_type,destination_name,destination_iban,destination_type,reconciled,category,budget,bill,tags,notes,sepa_cc,sepa_ct_op,sepa_ct_id,sepa_db,sepa_country,sepa_ep,sepa_ci,sepa_batch_id,external_url,interest_date,book_date,process_date,due_date,payment_date,invoice_date,recurrence_id,internal_reference,bunq_payment_id,import_hash,import_hash_v2,external_id,original_source,recurrence_total,recurrence_count,recurrence_date
1,2,3,4,5,6,Withdrawal,100,50,USD,PLN,desc,2024-06-27T12:00:00+00:00,UnknownAccount,notes,normal,17,18,19,category,21,22,tag1,notes2,extra`
		_, err := importer.Import(context.TODO(), &importers.ImportRequest{
			Data:     []byte(csv),
			Accounts: accountsData,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source account not found")
	})

	t.Run("currency mismatch", func(t *testing.T) {
		// Withdrawal with wrong currency
		csv := `user_id,group_id,journal_id,created_at,updated_at,group_title,type,amount,foreign_amount,currency_code,foreign_currency_code,description,date,source_name,source_iban,source_type,destination_name,destination_iban,destination_type,reconciled,category,budget,bill,tags,notes,sepa_cc,sepa_ct_op,sepa_ct_id,sepa_db,sepa_country,sepa_ep,sepa_ci,sepa_batch_id,external_url,interest_date,book_date,process_date,due_date,payment_date,invoice_date,recurrence_id,internal_reference,bunq_payment_id,import_hash,import_hash_v2,external_id,original_source,recurrence_total,recurrence_count,recurrence_date
1,2,3,4,5,6,Withdrawal,100,50,PLN,PLN,desc,2024-06-27T12:00:00+00:00,` + accountsData[0].Name + `,notes,normal,17,18,19,category,21,22,tag1,notes2,extra`
		_, err := importer.Import(context.TODO(), &importers.ImportRequest{
			Data:     []byte(csv),
			Accounts: accountsData,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "source account currency")
	})

	t.Run("unsupported operation type", func(t *testing.T) {
		csv := `user_id,group_id,journal_id,created_at,updated_at,group_title,type,amount,foreign_amount,currency_code,foreign_currency_code,description,date,source_name,source_iban,source_type,destination_name,destination_iban,destination_type,reconciled,category,budget,bill,tags,notes,sepa_cc,sepa_ct_op,sepa_ct_id,sepa_db,sepa_country,sepa_ep,sepa_ci,sepa_batch_id,external_url,interest_date,book_date,process_date,due_date,payment_date,invoice_date,recurrence_id,internal_reference,bunq_payment_id,import_hash,import_hash_v2,external_id,original_source,recurrence_total,recurrence_count,recurrence_date
1,2,3,4,5,6,UnknownType,100,50,USD,PLN,desc,2024-06-27T12:00:00+00:00,` + accountsData[0].Name + `,notes,normal,17,18,19,category,21,22,tag1,notes2,extra`
		_, err := importer.Import(context.TODO(), &importers.ImportRequest{
			Data:     []byte(csv),
			Accounts: accountsData,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported operation type")
	})

	t.Run("invalid amount", func(t *testing.T) {
		csv := `user_id,group_id,journal_id,created_at,updated_at,group_title,type,amount,foreign_amount,currency_code,foreign_currency_code,description,date,source_name,source_iban,source_type,destination_name,destination_iban,destination_type,reconciled,category,budget,bill,tags,notes,sepa_cc,sepa_ct_op,sepa_ct_id,sepa_db,sepa_country,sepa_ep,sepa_ci,sepa_batch_id,external_url,interest_date,book_date,process_date,due_date,payment_date,invoice_date,recurrence_id,internal_reference,bunq_payment_id,import_hash,import_hash_v2,external_id,original_source,recurrence_total,recurrence_count,recurrence_date
1,2,3,4,5,6,Withdrawal,notanumber,50,USD,PLN,desc,2024-06-27T12:00:00+00:00,` + accountsData[0].Name + `,notes,normal,17,18,19,category,21,22,tag1,notes2,extra`
		_, err := importer.Import(context.TODO(), &importers.ImportRequest{
			Data:     []byte(csv),
			Accounts: accountsData,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse amount")
	})
}

func TestParseDate(t *testing.T) {
	input := "2025-06-17T15:07:46+02:00"
	ff := importers.NewFireflyImporter(nil)

	t.Run("with local", func(t *testing.T) {
		resp, err := ff.ParseDate(input, false)
		assert.NoError(t, err)
		assert.Equal(t, input, resp.Format(time.RFC3339))
	})

	t.Run("force utc", func(t *testing.T) {
		resp, err := ff.ParseDate(input, true)
		assert.NoError(t, err)
		assert.Equal(t, "2025-06-17T15:07:46Z", resp.Format(time.RFC3339))
	})
}
