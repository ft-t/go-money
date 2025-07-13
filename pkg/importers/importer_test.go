package importers_test

import (
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"context"
	"encoding/base64"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImport(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		impl1 := NewMockImplementation(gomock.NewController(t))
		impl1.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)
		tag1 := NewMockTagSvc(gomock.NewController(t))

		targetTags := []*database.Tag{
			{
				Name: "ab",
				ID:   5,
			},
		}

		tag1.EXPECT().GetAllTags(gomock.Any()).Return(targetTags, nil)

		categoriesSvc := NewMockCategoriesSvc(gomock.NewController(t))
		categoriesSvc.EXPECT().GetAllCategories(gomock.Any()).Return([]*database.Category{
			{
				ID:   1,
				Name: "Category1",
			},
		}, nil)

		imp := importers.NewImporter(accSvc, tag1, categoriesSvc, impl1)

		rawBytes := []byte{0x1, 0x2}
		accounts := []*database.Account{
			{},
			{},
		}

		accSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)
		finalResp := &importv1.ImportTransactionsResponse{}

		impl1.EXPECT().Import(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, request *importers.ImportRequest) (*importv1.ImportTransactionsResponse, error) {
				assert.Equal(t, accounts, request.Accounts)
				assert.EqualValues(t, rawBytes, request.Data)

				assert.Len(t, request.Tags, 1)
				assert.EqualValues(t, 5, request.Tags["ab"].ID)

				return finalResp, nil
			})

		resp, err := imp.Import(context.TODO(), &importv1.ImportTransactionsRequest{
			FileContent: base64.StdEncoding.EncodeToString(rawBytes),
			Source:      importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
		})
		assert.NoError(t, err)
		assert.Equal(t, finalResp, resp)
	})

	t.Run("invalid source", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		impl1 := NewMockImplementation(gomock.NewController(t))
		impl1.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)

		tag1 := NewMockTagSvc(gomock.NewController(t))
		categoriesSvc := NewMockCategoriesSvc(gomock.NewController(t))

		imp := importers.NewImporter(accSvc, tag1, categoriesSvc, impl1)

		resp, err := imp.Import(context.TODO(), &importv1.ImportTransactionsRequest{
			FileContent: "test",
			Source:      importv1.ImportSource_IMPORT_SOURCE_UNSPECIFIED,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "unsupported import source")
	})

	t.Run("error from GetAllAccounts", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		impl1 := NewMockImplementation(gomock.NewController(t))
		impl1.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)

		tag1 := NewMockTagSvc(gomock.NewController(t))
		categoriesSvc := NewMockCategoriesSvc(gomock.NewController(t))

		imp := importers.NewImporter(accSvc, tag1, categoriesSvc, impl1)

		rawBytes := []byte{0x1, 0x2}

		accSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(nil, assert.AnError)

		resp, err := imp.Import(context.TODO(), &importv1.ImportTransactionsRequest{
			FileContent: base64.StdEncoding.EncodeToString(rawBytes),
			Source:      importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.EqualError(t, err, "failed to get accounts: "+assert.AnError.Error())
	})

	t.Run("error on decode base64", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		impl1 := NewMockImplementation(gomock.NewController(t))
		impl1.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)

		tag1 := NewMockTagSvc(gomock.NewController(t))
		categoriesSvc := NewMockCategoriesSvc(gomock.NewController(t))

		imp := importers.NewImporter(accSvc, tag1, categoriesSvc, impl1)

		resp, err := imp.Import(context.TODO(), &importv1.ImportTransactionsRequest{
			FileContent: "invalid_base64",
			Source:      importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "failed to decode file content: illegal base64 data at input byte")
	})

	t.Run("error on import", func(t *testing.T) {
		accSvc := NewMockAccountSvc(gomock.NewController(t))

		impl1 := NewMockImplementation(gomock.NewController(t))
		impl1.EXPECT().Type().Return(importv1.ImportSource_IMPORT_SOURCE_FIREFLY)

		tag1 := NewMockTagSvc(gomock.NewController(t))
		tag1.EXPECT().GetAllTags(gomock.Any()).Return([]*database.Tag{}, nil)
		categoriesSvc := NewMockCategoriesSvc(gomock.NewController(t))
		categoriesSvc.EXPECT().GetAllCategories(gomock.Any()).Return([]*database.Category{}, nil)

		imp := importers.NewImporter(accSvc, tag1, categoriesSvc, impl1)

		rawBytes := []byte{0x1, 0x2}
		accounts := []*database.Account{
			{},
			{},
		}

		accSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil)

		expectedErr := assert.AnError
		impl1.EXPECT().Import(gomock.Any(), gomock.Any()).
			Return(nil, expectedErr)

		resp, err := imp.Import(context.TODO(), &importv1.ImportTransactionsRequest{
			FileContent: base64.StdEncoding.EncodeToString(rawBytes),
			Source:      importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
		})
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "import failed")
	})
}
