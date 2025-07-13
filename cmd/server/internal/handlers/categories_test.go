package handlers_test

import (
	categoriesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/categories/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func newCategoriesApiWithMock(t *testing.T) (*handlers.CategoriesApi, *MockCategoriesSvc) {
	ctrl := gomock.NewController(t)
	catSvc := NewMockCategoriesSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api := handlers.NewCategoriesApi(grpc, catSvc)
	return api, catSvc
}

func TestCategoriesApi_CreateCategory(t *testing.T) {
	api, catSvc := newCategoriesApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&categoriesv1.CreateCategoryRequest{})
		respMsg := &categoriesv1.CreateCategoryResponse{}
		catSvc.EXPECT().CreateCategory(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.CreateCategory(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&categoriesv1.CreateCategoryRequest{})
		catSvc.EXPECT().CreateCategory(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.CreateCategory(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&categoriesv1.CreateCategoryRequest{})
		resp, err := api.CreateCategory(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestCategoriesApi_UpdateCategory(t *testing.T) {
	api, catSvc := newCategoriesApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&categoriesv1.UpdateCategoryRequest{})
		respMsg := &categoriesv1.UpdateCategoryResponse{}
		catSvc.EXPECT().UpdateCategory(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.UpdateCategory(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&categoriesv1.UpdateCategoryRequest{})
		catSvc.EXPECT().UpdateCategory(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.UpdateCategory(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&categoriesv1.UpdateCategoryRequest{})
		resp, err := api.UpdateCategory(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestCategoriesApi_DeleteCategory(t *testing.T) {
	api, catSvc := newCategoriesApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&categoriesv1.DeleteCategoryRequest{})
		catSvc.EXPECT().DeleteCategory(gomock.Any(), req.Msg).Return(&categoriesv1.DeleteCategoryResponse{}, nil)
		resp, err := api.DeleteCategory(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&categoriesv1.DeleteCategoryRequest{})
		catSvc.EXPECT().DeleteCategory(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.DeleteCategory(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&categoriesv1.DeleteCategoryRequest{})
		resp, err := api.DeleteCategory(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestCategoriesApi_ListCategories(t *testing.T) {
	api, catSvc := newCategoriesApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&categoriesv1.ListCategoriesRequest{})
		respMsg := &categoriesv1.ListCategoriesResponse{}
		catSvc.EXPECT().ListCategories(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.ListCategories(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&categoriesv1.ListCategoriesRequest{})
		catSvc.EXPECT().ListCategories(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.ListCategories(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&categoriesv1.ListCategoriesRequest{})
		resp, err := api.ListCategories(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}
