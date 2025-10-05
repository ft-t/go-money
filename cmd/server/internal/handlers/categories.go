package handlers

import (
	"context"

	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/categories/v1/categoriesv1connect"
	categoriesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/categories/v1"
	"connectrpc.com/connect"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type CategoriesApi struct {
	categoriesSvc CategoriesSvc
}

func NewCategoriesApi(
	mux *boilerplate.DefaultGrpcServer,
	categoriesSvc CategoriesSvc,
) *CategoriesApi {
	res := &CategoriesApi{
		categoriesSvc: categoriesSvc,
	}

	mux.GetMux().Handle(
		categoriesv1connect.NewCategoriesServiceHandler(res, mux.GetDefaultHandlerOptions()...),
	)

	return res
}

func (c *CategoriesApi) CreateCategory(ctx context.Context, req *connect.Request[categoriesv1.CreateCategoryRequest]) (*connect.Response[categoriesv1.CreateCategoryResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := c.categoriesSvc.CreateCategory(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}

func (c *CategoriesApi) UpdateCategory(ctx context.Context, req *connect.Request[categoriesv1.UpdateCategoryRequest]) (*connect.Response[categoriesv1.UpdateCategoryResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := c.categoriesSvc.UpdateCategory(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}

func (c *CategoriesApi) DeleteCategory(ctx context.Context, req *connect.Request[categoriesv1.DeleteCategoryRequest]) (*connect.Response[categoriesv1.DeleteCategoryResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := c.categoriesSvc.DeleteCategory(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}

func (c *CategoriesApi) ListCategories(ctx context.Context, req *connect.Request[categoriesv1.ListCategoriesRequest]) (*connect.Response[categoriesv1.ListCategoriesResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := c.categoriesSvc.ListCategories(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}
