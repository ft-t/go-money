package categories_test

import (
	categoriesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/categories/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/categories"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

var gormDB *gorm.DB
var cfg *configuration.Configuration

func TestMain(m *testing.M) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestCreateCategory(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapper(gomock.NewController(t))
		srv := categories.NewService(mapper)

		mapper.EXPECT().MapCategory(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, category *database.Category) *gomoneypbv1.Category {
				assert.NotEmpty(t, category.ID)

				return &gomoneypbv1.Category{}
			})

		resp, err := srv.CreateCategory(context.TODO(), &categoriesv1.CreateCategoryRequest{
			Category: &gomoneypbv1.Category{
				Name: "Test Category",
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("duplicate name", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapper(gomock.NewController(t))
		srv := categories.NewService(mapper)

		category := &database.Category{
			Name: "Duplicate Category",
		}
		assert.NoError(t, gormDB.Create(category).Error)

		resp, err := srv.CreateCategory(context.TODO(), &categoriesv1.CreateCategoryRequest{
			Category: &gomoneypbv1.Category{
				Name: "Duplicate Category",
			},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "duplicated key not allowed")
	})

	t.Run("no exist err", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapper(gomock.NewController(t))
		srv := categories.NewService(mapper)

		category := &database.Category{
			Name: "Duplicate Category",
		}
		assert.NoError(t, gormDB.Create(category).Error)

		ctx, cancel := context.WithCancel(context.TODO())
		cancel()
		resp, err := srv.CreateCategory(ctx, &categoriesv1.CreateCategoryRequest{
			Category: &gomoneypbv1.Category{
				Name: "Duplicate Category",
			},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestUpdateCategory(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapper(gomock.NewController(t))
		srv := categories.NewService(mapper)

		category := &database.Category{
			Name: "Initial Category",
		}
		assert.NoError(t, gormDB.Create(category).Error)

		mapper.EXPECT().MapCategory(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, category *database.Category) *gomoneypbv1.Category {
				assert.EqualValues(t, category.Name, "Updated Category")
				return &gomoneypbv1.Category{}
			})

		resp, err := srv.UpdateCategory(context.TODO(), &categoriesv1.UpdateCategoryRequest{
			Category: &gomoneypbv1.Category{
				Id:   category.ID,
				Name: "Updated Category",
			},
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("not found", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapper(gomock.NewController(t))
		srv := categories.NewService(mapper)

		resp, err := srv.UpdateCategory(context.TODO(), &categoriesv1.UpdateCategoryRequest{
			Category: &gomoneypbv1.Category{
				Id:   9999,
				Name: "Non-existent Category",
			},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "record not found")
	})

	t.Run("duplicate name", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapper(gomock.NewController(t))
		srv := categories.NewService(mapper)

		category1 := &database.Category{
			Name: "Category One",
		}
		assert.NoError(t, gormDB.Create(category1).Error)

		category2 := &database.Category{
			Name: "Category Two",
		}
		assert.NoError(t, gormDB.Create(category2).Error)

		resp, err := srv.UpdateCategory(context.TODO(), &categoriesv1.UpdateCategoryRequest{
			Category: &gomoneypbv1.Category{
				Id:   category2.ID,
				Name: "Category One",
			},
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "duplicated key not allowed")
	})
}

func TestDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapper(gomock.NewController(t))
		srv := categories.NewService(mapper)

		mapper.EXPECT().MapCategory(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, category *database.Category) *gomoneypbv1.Category {
				assert.True(t, category.DeletedAt.Valid)

				return &gomoneypbv1.Category{}
			})

		category := &database.Category{
			Name: "Category to Delete",
		}
		assert.NoError(t, gormDB.Create(category).Error)

		resp, err := srv.DeleteCategory(context.TODO(), &categoriesv1.DeleteCategoryRequest{
			Id: category.ID,
		})

		assert.NoError(t, err)
		assert.NotNil(t, resp)

		var deletedCategory database.Category

		assert.ErrorIs(t, gormDB.Where("id = ?", category.ID).First(&deletedCategory).Error,
			gorm.ErrRecordNotFound)
	})

	t.Run("not found", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		mapper := NewMockMapper(gomock.NewController(t))
		srv := categories.NewService(mapper)

		resp, err := srv.DeleteCategory(context.TODO(), &categoriesv1.DeleteCategoryRequest{
			Id: 9999,
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.ErrorContains(t, err, "record not found")
	})
}

func TestListing(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	cats := []*database.Category{
		{
			Name: "Category 1",
		},
		{
			Name: "Category 2",
			DeletedAt: gorm.DeletedAt{
				Time:  time.Now(),
				Valid: true,
			},
		},
	}
	assert.NoError(t, gormDB.Create(&cats).Error)

	t.Run("include deleted", func(t *testing.T) {
		mapper := NewMockMapper(gomock.NewController(t))
		srv := categories.NewService(mapper)

		mapper.EXPECT().MapCategory(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, category *database.Category) *gomoneypbv1.Category {
				return &gomoneypbv1.Category{
					Id: category.ID,
				}
			}).Times(2)

		resp, err := srv.ListCategories(context.TODO(), &categoriesv1.ListCategoriesRequest{
			IncludeDeleted: true,
		})

		assert.NoError(t, err)
		assert.Len(t, resp.Categories, 2)
	})

	t.Run("with ids", func(t *testing.T) {
		mapper := NewMockMapper(gomock.NewController(t))
		srv := categories.NewService(mapper)

		mapper.EXPECT().MapCategory(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, category *database.Category) *gomoneypbv1.Category {
				return &gomoneypbv1.Category{
					Id: category.ID,
				}
			})

		resp, err := srv.ListCategories(context.TODO(), &categoriesv1.ListCategoriesRequest{
			Ids: []int32{cats[0].ID},
		})

		assert.NoError(t, err)
		assert.Len(t, resp.Categories, 1)
		assert.EqualValues(t, cats[0].ID, resp.Categories[0].Id)
	})
}
