package handlers_test

import (
	tagsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/tags/v1"
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

func newTagsApiWithMock(t *testing.T) (*handlers.TagsApi, *MockTagSvc) {
	ctrl := gomock.NewController(t)
	tagSvc := NewMockTagSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api := handlers.NewTagsApi(grpc, tagSvc)
	return api, tagSvc
}

func TestTagsApi_CreateTag(t *testing.T) {
	api, tagSvc := newTagsApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&tagsv1.CreateTagRequest{})
		respMsg := &tagsv1.CreateTagResponse{}
		tagSvc.EXPECT().CreateTag(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.CreateTag(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&tagsv1.CreateTagRequest{})
		tagSvc.EXPECT().CreateTag(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.CreateTag(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&tagsv1.CreateTagRequest{})
		resp, err := api.CreateTag(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestTagsApi_ImportTags(t *testing.T) {
	api, tagSvc := newTagsApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&tagsv1.ImportTagsRequest{})
		respMsg := &tagsv1.ImportTagsResponse{}
		tagSvc.EXPECT().ImportTags(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.ImportTags(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&tagsv1.ImportTagsRequest{})
		tagSvc.EXPECT().ImportTags(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.ImportTags(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&tagsv1.ImportTagsRequest{})
		resp, err := api.ImportTags(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestTagsApi_UpdateTag(t *testing.T) {
	api, tagSvc := newTagsApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&tagsv1.UpdateTagRequest{})
		respMsg := &tagsv1.UpdateTagResponse{}
		tagSvc.EXPECT().UpdateTag(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.UpdateTag(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&tagsv1.UpdateTagRequest{})
		tagSvc.EXPECT().UpdateTag(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.UpdateTag(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&tagsv1.UpdateTagRequest{})
		resp, err := api.UpdateTag(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestTagsApi_DeleteTag(t *testing.T) {
	api, tagSvc := newTagsApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&tagsv1.DeleteTagRequest{})
		tagSvc.EXPECT().DeleteTag(gomock.Any(), req.Msg).Return(nil)
		resp, err := api.DeleteTag(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&tagsv1.DeleteTagRequest{})
		tagSvc.EXPECT().DeleteTag(gomock.Any(), req.Msg).Return(assert.AnError)
		resp, err := api.DeleteTag(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&tagsv1.DeleteTagRequest{})
		resp, err := api.DeleteTag(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestTagsApi_ListTags(t *testing.T) {
	api, tagSvc := newTagsApiWithMock(t)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&tagsv1.ListTagsRequest{})
		respMsg := &tagsv1.ListTagsResponse{}
		tagSvc.EXPECT().ListTags(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.ListTags(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&tagsv1.ListTagsRequest{})
		tagSvc.EXPECT().ListTags(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.ListTags(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&tagsv1.ListTagsRequest{})
		resp, err := api.ListTags(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}
