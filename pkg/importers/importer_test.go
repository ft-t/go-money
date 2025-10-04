package importers_test

import (
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
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
					Title:                   "Test Transaction",
					InternalReferenceNumber: strPtr("test_ref_123"),
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
					Title:                   "Test Transaction",
					InternalReferenceNumber: strPtr("test_ref_123"),
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
					Title:                   "Test Transaction",
					InternalReferenceNumber: strPtr("ref123"),
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
}

func strPtr(s string) *string {
	return &s
}
