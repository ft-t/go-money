package importers_test

import (
	"context"
	"testing"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/ft-t/go-money/pkg/transactions/history"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestImport_SetsImporterActor(t *testing.T) {
	cases := []struct {
		name       string
		source     importv1.ImportSource
		expectName string
	}{
		{"firefly", importv1.ImportSource_IMPORT_SOURCE_FIREFLY, "firefly"},
		{"privat24", importv1.ImportSource_IMPORT_SOURCE_PRIVATE_24, "privat24"},
		{"revolut", importv1.ImportSource_IMPORT_SOURCE_REVOLUT, "revolut"},
		{"monobank", importv1.ImportSource_IMPORT_SOURCE_MONOBANK, "monobank"},
		{"paribas", importv1.ImportSource_IMPORT_SOURCE_BNP_PARIBAS_POLSKA, "paribas"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			accSvc := NewMockAccountSvc(ctrl)
			tagSvc := NewMockTagSvc(ctrl)
			categoriesSvc := NewMockCategoriesSvc(ctrl)
			txSvc := NewMockTransactionSvc(ctrl)
			mapperSvc := NewMockMapperSvc(ctrl)

			impl := NewMockImplementation(ctrl)
			impl.EXPECT().Type().Return(tc.source)

			cfg := &importers.ImporterConfig{
				AccountSvc:     accSvc,
				TagSvc:         tagSvc,
				CategoriesSvc:  categoriesSvc,
				TransactionSvc: txSvc,
				MapperSvc:      mapperSvc,
			}

			imp := importers.NewImporter(cfg, impl)

			accSvc.EXPECT().GetAllAccounts(gomock.Any()).Return([]*database.Account{{ID: 1}}, nil)
			tagSvc.EXPECT().GetAllTags(gomock.Any()).Return([]*database.Tag{}, nil)
			categoriesSvc.EXPECT().GetAllCategories(gomock.Any()).Return([]*database.Category{}, nil)

			parseResp := &importers.ParseResponse{
				CreateRequests: []*transactionsv1.CreateTransactionRequest{
					{
						Title:                    "Test",
						InternalReferenceNumbers: []string{"ref_1"},
					},
				},
			}
			impl.EXPECT().Parse(gomock.Any(), gomock.Any()).Return(parseResp, nil)

			txSvc.EXPECT().
				CreateBulkInternal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(
					ctx context.Context,
					_ []*transactions.BulkRequest,
					_ *gorm.DB,
					_ transactions.UpsertOptions,
				) ([]*transactionsv1.CreateTransactionResponse, error) {
					actor, ok := history.ActorFromContext(ctx)
					require.True(t, ok, "history actor must be set in context")
					assert.Equal(t, database.TransactionHistoryActorTypeImporter, actor.Type)
					assert.Equal(t, tc.expectName, actor.Detail)
					return []*transactionsv1.CreateTransactionResponse{
						{Transaction: &gomoneypbv1.Transaction{Id: 1}},
					}, nil
				})

			resp, err := imp.Import(context.TODO(), &importv1.ImportTransactionsRequest{
				Content: []string{"test content"},
				Source:  tc.source,
			})
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.EqualValues(t, 1, resp.ImportedCount)
		})
	}
}

func TestImport_UnknownSourceNameFallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	accSvc := NewMockAccountSvc(ctrl)
	tagSvc := NewMockTagSvc(ctrl)
	categoriesSvc := NewMockCategoriesSvc(ctrl)
	txSvc := NewMockTransactionSvc(ctrl)
	mapperSvc := NewMockMapperSvc(ctrl)

	impl := NewMockImplementation(ctrl)
	impl.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_UNSPECIFIED)

	cfg := &importers.ImporterConfig{
		AccountSvc:     accSvc,
		TagSvc:         tagSvc,
		CategoriesSvc:  categoriesSvc,
		TransactionSvc: txSvc,
		MapperSvc:      mapperSvc,
	}

	imp := importers.NewImporter(cfg, impl)

	accSvc.EXPECT().GetAllAccounts(gomock.Any()).Return([]*database.Account{{ID: 1}}, nil)
	tagSvc.EXPECT().GetAllTags(gomock.Any()).Return([]*database.Tag{}, nil)
	categoriesSvc.EXPECT().GetAllCategories(gomock.Any()).Return([]*database.Category{}, nil)

	parseResp := &importers.ParseResponse{
		CreateRequests: []*transactionsv1.CreateTransactionRequest{
			{
				Title:                    "Test",
				InternalReferenceNumbers: []string{"ref_x"},
			},
		},
	}
	impl.EXPECT().Parse(gomock.Any(), gomock.Any()).Return(parseResp, nil)

	txSvc.EXPECT().
		CreateBulkInternal(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(
			ctx context.Context,
			_ []*transactions.BulkRequest,
			_ *gorm.DB,
			_ transactions.UpsertOptions,
		) ([]*transactionsv1.CreateTransactionResponse, error) {
			actor, ok := history.ActorFromContext(ctx)
			require.True(t, ok)
			assert.Equal(t, database.TransactionHistoryActorTypeImporter, actor.Type)
			assert.Equal(t, "unknown", actor.Detail)
			return []*transactionsv1.CreateTransactionResponse{
				{Transaction: &gomoneypbv1.Transaction{Id: 1}},
			}, nil
		})

	resp, err := imp.Import(context.TODO(), &importv1.ImportTransactionsRequest{
		Content: []string{"test content"},
		Source:  importv1.ImportSource_IMPORT_SOURCE_UNSPECIFIED,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}
