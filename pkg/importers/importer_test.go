package importers_test

import (
	"context"
	"testing"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestImport(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		accSvc := NewMockAccountSvc(ctrl)
		tagSvc := NewMockTagSvc(ctrl)
		categoriesSvc := NewMockCategoriesSvc(ctrl)
		txSvc := NewMockTransactionSvc(ctrl)
		mapperSvc := NewMockMapperSvc(ctrl)

		impl1 := NewMockImplementation(ctrl)
		impl1.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)

		cfg := &importers.ImporterConfig{
			AccountSvc:     accSvc,
			TagSvc:         tagSvc,
			CategoriesSvc:  categoriesSvc,
			TransactionSvc: txSvc,
			MapperSvc:      mapperSvc,
		}

		imp := importers.NewImporter(cfg, impl1)

		targetTags := []*database.Tag{{Name: "ab", ID: 5}}
		targetCategories := []*database.Category{{ID: 1, Name: "Category1"}}
		accounts := []*database.Account{{ID: 1}, {ID: 2}}

		tagSvc.EXPECT().GetAllTags(gomock.Any()).Return(targetTags, nil)
		categoriesSvc.EXPECT().GetAllCategories(gomock.Any()).Return(targetCategories, nil)
		accSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)

		parseResp := &importers.ParseResponse{
			CreateRequests: []*transactionsv1.CreateTransactionRequest{
				{
					Title:                    "Test Transaction",
					InternalReferenceNumbers: []string{"test_ref_123"},
				},
			},
		}

		impl1.EXPECT().Parse(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, request *importers.ParseRequest) (*importers.ParseResponse, error) {
				assert.Equal(t, accounts, request.Accounts)
				assert.Equal(t, []string{"test content"}, request.Data)
				assert.Len(t, request.Tags, 1)
				assert.EqualValues(t, 5, request.Tags["ab"].ID)
				return parseResp, nil
			})

		txSvc.EXPECT().CreateBulkInternal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return([]*transactionsv1.CreateTransactionResponse{
				{Transaction: &gomoneypbv1.Transaction{Id: 123}},
			}, nil)

		resp, err := imp.Import(context.TODO(), &importv1.ImportTransactionsRequest{
			Content: []string{"test content"},
			Source:  importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.EqualValues(t, 1, resp.ImportedCount)
	})

	t.Run("invalid source", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		impl1 := NewMockImplementation(ctrl)
		impl1.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)

		cfg := &importers.ImporterConfig{}
		imp := importers.NewImporter(cfg, impl1)

		resp, err := imp.Import(context.TODO(), &importv1.ImportTransactionsRequest{
			Content: []string{"test"},
			Source:  importv1.ImportSource_IMPORT_SOURCE_UNSPECIFIED,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "unsupported import source")
	})

	t.Run("error from GetAllAccounts", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		accSvc := NewMockAccountSvc(ctrl)
		tagSvc := NewMockTagSvc(ctrl)
		categoriesSvc := NewMockCategoriesSvc(ctrl)
		txSvc := NewMockTransactionSvc(ctrl)
		mapperSvc := NewMockMapperSvc(ctrl)

		impl1 := NewMockImplementation(ctrl)
		impl1.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)

		cfg := &importers.ImporterConfig{
			AccountSvc:     accSvc,
			TagSvc:         tagSvc,
			CategoriesSvc:  categoriesSvc,
			TransactionSvc: txSvc,
			MapperSvc:      mapperSvc,
		}

		imp := importers.NewImporter(cfg, impl1)

		accSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(nil, assert.AnError)

		resp, err := imp.Import(context.TODO(), &importv1.ImportTransactionsRequest{
			Content: []string{"test"},
			Source:  importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to get accounts")
	})

	t.Run("error on parse", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		accSvc := NewMockAccountSvc(ctrl)
		tagSvc := NewMockTagSvc(ctrl)
		categoriesSvc := NewMockCategoriesSvc(ctrl)
		txSvc := NewMockTransactionSvc(ctrl)
		mapperSvc := NewMockMapperSvc(ctrl)

		impl1 := NewMockImplementation(ctrl)
		impl1.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)

		cfg := &importers.ImporterConfig{
			AccountSvc:     accSvc,
			TagSvc:         tagSvc,
			CategoriesSvc:  categoriesSvc,
			TransactionSvc: txSvc,
			MapperSvc:      mapperSvc,
		}

		imp := importers.NewImporter(cfg, impl1)

		accounts := []*database.Account{{ID: 1}, {ID: 2}}

		tagSvc.EXPECT().GetAllTags(gomock.Any()).Return([]*database.Tag{}, nil)
		categoriesSvc.EXPECT().GetAllCategories(gomock.Any()).Return([]*database.Category{}, nil)
		accSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)

		impl1.EXPECT().Parse(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)

		resp, err := imp.Import(context.TODO(), &importv1.ImportTransactionsRequest{
			Content: []string{"test"},
			Source:  importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("error on create bulk", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		accSvc := NewMockAccountSvc(ctrl)
		tagSvc := NewMockTagSvc(ctrl)
		categoriesSvc := NewMockCategoriesSvc(ctrl)
		txSvc := NewMockTransactionSvc(ctrl)
		mapperSvc := NewMockMapperSvc(ctrl)

		impl1 := NewMockImplementation(ctrl)
		impl1.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)

		cfg := &importers.ImporterConfig{
			AccountSvc:     accSvc,
			TagSvc:         tagSvc,
			CategoriesSvc:  categoriesSvc,
			TransactionSvc: txSvc,
			MapperSvc:      mapperSvc,
		}

		imp := importers.NewImporter(cfg, impl1)

		accounts := []*database.Account{{ID: 1}, {ID: 2}}
		parseResp := &importers.ParseResponse{
			CreateRequests: []*transactionsv1.CreateTransactionRequest{
				{
					Title:                    "Test Transaction",
					InternalReferenceNumbers: []string{"test_ref_123"},
				},
			},
		}

		tagSvc.EXPECT().GetAllTags(gomock.Any()).Return([]*database.Tag{}, nil)
		categoriesSvc.EXPECT().GetAllCategories(gomock.Any()).Return([]*database.Category{}, nil)
		accSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)
		impl1.EXPECT().Parse(gomock.Any(), gomock.Any()).Return(parseResp, nil)
		txSvc.EXPECT().CreateBulkInternal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, assert.AnError)

		resp, err := imp.Import(context.TODO(), &importv1.ImportTransactionsRequest{
			Content: []string{"test"},
			Source:  importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestParse(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		accSvc := NewMockAccountSvc(ctrl)
		tagSvc := NewMockTagSvc(ctrl)
		categoriesSvc := NewMockCategoriesSvc(ctrl)
		txSvc := NewMockTransactionSvc(ctrl)
		mapperSvc := NewMockMapperSvc(ctrl)

		impl1 := NewMockImplementation(ctrl)
		impl1.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)

		cfg := &importers.ImporterConfig{
			AccountSvc:     accSvc,
			TagSvc:         tagSvc,
			CategoriesSvc:  categoriesSvc,
			TransactionSvc: txSvc,
			MapperSvc:      mapperSvc,
		}

		imp := importers.NewImporter(cfg, impl1)

		targetTags := []*database.Tag{{Name: "ab", ID: 5}}
		targetCategories := []*database.Category{{ID: 1, Name: "Category1"}}
		accounts := []*database.Account{{ID: 1}, {ID: 2}}

		tagSvc.EXPECT().GetAllTags(gomock.Any()).Return(targetTags, nil)
		categoriesSvc.EXPECT().GetAllCategories(gomock.Any()).Return(targetCategories, nil)
		accSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)

		parseResp := &importers.ParseResponse{
			CreateRequests: []*transactionsv1.CreateTransactionRequest{
				{
					Title:                    "Test Transaction",
					InternalReferenceNumbers: []string{"ref123"},
					Transaction: &transactionsv1.CreateTransactionRequest_Expense{
						Expense: &transactionsv1.Expense{
							SourceAccountId:      1,
							SourceAmount:         "-100",
							SourceCurrency:       "USD",
							DestinationAccountId: 2,
							DestinationAmount:    "100",
							DestinationCurrency:  "USD",
						},
					},
				},
			},
		}

		impl1.EXPECT().Parse(gomock.Any(), gomock.Any()).Return(parseResp, nil)
		txSvc.EXPECT().ConvertRequestToTransaction(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&database.Transaction{ID: 1}, nil)
		mapperSvc.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).Return(&gomoneypbv1.Transaction{Id: 1})

		resp, err := imp.Parse(context.TODO(), &importv1.ParseTransactionsRequest{
			Content: []string{"test content"},
			Source:  importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Transactions, 1)
	})

	t.Run("parse error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		accSvc := NewMockAccountSvc(ctrl)
		tagSvc := NewMockTagSvc(ctrl)
		categoriesSvc := NewMockCategoriesSvc(ctrl)
		txSvc := NewMockTransactionSvc(ctrl)
		mapperSvc := NewMockMapperSvc(ctrl)

		impl1 := NewMockImplementation(ctrl)
		impl1.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)

		cfg := &importers.ImporterConfig{
			AccountSvc:     accSvc,
			TagSvc:         tagSvc,
			CategoriesSvc:  categoriesSvc,
			TransactionSvc: txSvc,
			MapperSvc:      mapperSvc,
		}

		imp := importers.NewImporter(cfg, impl1)

		accounts := []*database.Account{{ID: 1}}
		tagSvc.EXPECT().GetAllTags(gomock.Any()).Return([]*database.Tag{}, nil)
		categoriesSvc.EXPECT().GetAllCategories(gomock.Any()).Return([]*database.Category{}, nil)
		accSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)
		impl1.EXPECT().Parse(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)

		resp, err := imp.Parse(context.TODO(), &importv1.ParseTransactionsRequest{
			Content: []string{"test"},
			Source:  importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("parse with failed transaction (no transaction set)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		accSvc := NewMockAccountSvc(ctrl)
		tagSvc := NewMockTagSvc(ctrl)
		categoriesSvc := NewMockCategoriesSvc(ctrl)
		txSvc := NewMockTransactionSvc(ctrl)
		mapperSvc := NewMockMapperSvc(ctrl)

		impl1 := NewMockImplementation(ctrl)
		impl1.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)

		cfg := &importers.ImporterConfig{
			AccountSvc:     accSvc,
			TagSvc:         tagSvc,
			CategoriesSvc:  categoriesSvc,
			TransactionSvc: txSvc,
			MapperSvc:      mapperSvc,
		}

		imp := importers.NewImporter(cfg, impl1)

		targetTags := []*database.Tag{{Name: "ab", ID: 5}}
		targetCategories := []*database.Category{{ID: 1, Name: "Category1"}}
		accounts := []*database.Account{{ID: 1}, {ID: 2}}

		tagSvc.EXPECT().GetAllTags(gomock.Any()).Return(targetTags, nil)
		categoriesSvc.EXPECT().GetAllCategories(gomock.Any()).Return(targetCategories, nil)
		accSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)

		parseResp := &importers.ParseResponse{
			CreateRequests: []*transactionsv1.CreateTransactionRequest{
				{
					Title:                    "Failed Transaction",
					Notes:                    "Raw transaction data",
					InternalReferenceNumbers: []string{"ref456"},
				},
			},
		}

		impl1.EXPECT().Parse(gomock.Any(), gomock.Any()).Return(parseResp, nil)
		mapperSvc.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, tx *database.Transaction) *gomoneypbv1.Transaction {
				assert.Equal(t, "Failed Transaction", tx.Title)
				assert.Equal(t, "Raw transaction data", tx.Notes)
				assert.Equal(t, gomoneypbv1.TransactionType_TRANSACTION_TYPE_UNSPECIFIED, tx.TransactionType)
				return &gomoneypbv1.Transaction{
					Title: tx.Title,
					Notes: tx.Notes,
					Type:  gomoneypbv1.TransactionType_TRANSACTION_TYPE_UNSPECIFIED,
				}
			})

		resp, err := imp.Parse(context.TODO(), &importv1.ParseTransactionsRequest{
			Content: []string{"test content"},
			Source:  importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Transactions, 1)
		assert.Equal(t, gomoneypbv1.TransactionType_TRANSACTION_TYPE_UNSPECIFIED, resp.Transactions[0].Transaction.Type)
		assert.Equal(t, "Failed Transaction", resp.Transactions[0].Transaction.Title)
	})
}

func TestCheckDuplicates(t *testing.T) {
	t.Run("success with no duplicates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		impl := NewMockImplementation(ctrl)
		impl.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)
		imp := importers.NewImporter(&importers.ImporterConfig{}, impl)

		requests := []*transactionsv1.CreateTransactionRequest{
			{
				Title:                    "Transaction 1",
				InternalReferenceNumbers: []string{"ref_001"},
			},
			{
				Title:                    "Transaction 2",
				InternalReferenceNumbers: []string{"ref_002"},
			},
		}

		result, err := imp.CheckDuplicates(context.TODO(), requests)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.NotNil(t, result["ref_001"])
		assert.NotNil(t, result["ref_002"])
		assert.Nil(t, result["ref_001"].DuplicationTransactionID)
		assert.Nil(t, result["ref_002"].DuplicationTransactionID)
	})

	t.Run("error when missing internal reference number", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		impl := NewMockImplementation(ctrl)
		impl.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)
		imp := importers.NewImporter(&importers.ImporterConfig{}, impl)

		requests := []*transactionsv1.CreateTransactionRequest{
			{
				Title:                    "Transaction 1",
				InternalReferenceNumbers: nil,
			},
		}

		result, err := imp.CheckDuplicates(context.TODO(), requests)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "all transactions must have at least one reference number for deduplication")
	})

	t.Run("error when empty internal reference number", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		impl := NewMockImplementation(ctrl)
		impl.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)
		imp := importers.NewImporter(&importers.ImporterConfig{}, impl)

		requests := []*transactionsv1.CreateTransactionRequest{
			{
				Title:                    "Transaction 1",
				InternalReferenceNumbers: []string{""},
			},
		}

		result, err := imp.CheckDuplicates(context.TODO(), requests)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "all transactions must have at least one reference number for deduplication")
	})

	t.Run("error when whitespace-only internal reference number", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		impl := NewMockImplementation(ctrl)
		impl.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)
		imp := importers.NewImporter(&importers.ImporterConfig{}, impl)

		requests := []*transactionsv1.CreateTransactionRequest{
			{
				Title:                    "Transaction 1",
				InternalReferenceNumbers: []string{"   "},
			},
		}

		result, err := imp.CheckDuplicates(context.TODO(), requests)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "all transactions must have at least one reference number for deduplication")
	})

	t.Run("error on duplicate reference in import data", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		impl := NewMockImplementation(ctrl)
		impl.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)
		imp := importers.NewImporter(&importers.ImporterConfig{}, impl)

		requests := []*transactionsv1.CreateTransactionRequest{
			{
				Title:                    "Transaction 1",
				InternalReferenceNumbers: []string{"ref_duplicate"},
			},
			{
				Title:                    "Transaction 2",
				InternalReferenceNumbers: []string{"ref_duplicate"},
			},
		}

		result, err := imp.CheckDuplicates(context.TODO(), requests)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "duplicate reference number found in import data: ref=ref_duplicate")
	})

	t.Run("db error on check", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockGorm, _, sql := testingutils.GormMock()

		impl := NewMockImplementation(ctrl)
		impl.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)
		imp := importers.NewImporter(&importers.ImporterConfig{}, impl)

		sql.ExpectQuery("SELECT internal_reference_numbers, id FROM \"transactions\"").
			WillReturnError(errors.New("failed to check existing transactions"))

		ctx := database.WithContext(context.TODO(), mockGorm)

		requests := []*transactionsv1.CreateTransactionRequest{
			{
				Title:                    "Transaction 1",
				InternalReferenceNumbers: []string{"ref_001"},
			},
		}

		result, err := imp.CheckDuplicates(ctx, requests)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to check existing transactions")
	})
}
