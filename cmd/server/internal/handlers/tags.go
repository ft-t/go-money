package handlers

import (
	"context"

	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/tags/v1/tagsv1connect"
	tagsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/tags/v1"
	"connectrpc.com/connect"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type TagsApi struct {
	tagSvc TagSvc
}

func NewTagsApi(
	mux *boilerplate.DefaultGrpcServer,
	tagSvc TagSvc,
) *TagsApi {
	res := &TagsApi{
		tagSvc: tagSvc,
	}

	mux.GetMux().Handle(
		tagsv1connect.NewTagsServiceHandler(res, mux.GetDefaultHandlerOptions()...),
	)

	return res
}

func (t *TagsApi) CreateTag(ctx context.Context, c *connect.Request[tagsv1.CreateTagRequest]) (*connect.Response[tagsv1.CreateTagResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := t.tagSvc.CreateTag(ctx, c.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}

func (t *TagsApi) ImportTags(ctx context.Context, c *connect.Request[tagsv1.ImportTagsRequest]) (*connect.Response[tagsv1.ImportTagsResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := t.tagSvc.ImportTags(ctx, c.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}

func (t *TagsApi) UpdateTag(ctx context.Context, c *connect.Request[tagsv1.UpdateTagRequest]) (*connect.Response[tagsv1.UpdateTagResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := t.tagSvc.UpdateTag(ctx, c.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}

func (t *TagsApi) DeleteTag(ctx context.Context, c *connect.Request[tagsv1.DeleteTagRequest]) (*connect.Response[tagsv1.DeleteTagResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	err := t.tagSvc.DeleteTag(ctx, c.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&tagsv1.DeleteTagResponse{
		Tag: nil,
	}), nil
}

func (t *TagsApi) ListTags(
	ctx context.Context,
	c *connect.Request[tagsv1.ListTagsRequest],
) (*connect.Response[tagsv1.ListTagsResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := t.tagSvc.ListTags(ctx, c.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}
