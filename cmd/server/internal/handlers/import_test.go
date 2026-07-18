package handlers_test

import (
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/importers"
	"github.com/golang/mock/gomock"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewImportApi(t *testing.T) {
	mockSvc := NewMockImportSvc(gomock.NewController(t))
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, err := handlers.NewImportApi(grpc, mockSvc)
	assert.NoError(t, err)
	assert.NotNil(t, api)
}

func TestImportApi_ImportTransactions(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockImportSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewImportApi(grpc, mockSvc)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&importv1.ImportTransactionsRequest{})
		respMsg := &importv1.ImportTransactionsResponse{}
		mockSvc.EXPECT().Import(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.ImportTransactions(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&importv1.ImportTransactionsRequest{})
		mockSvc.EXPECT().Import(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.ImportTransactions(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&importv1.ImportTransactionsRequest{})
		resp, err := api.ImportTransactions(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestImportApi_ParseTransactions(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockImportSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewImportApi(grpc, mockSvc)

	t.Run("success", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&importv1.ParseTransactionsRequest{
			Source:  importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
			Content: []string{"test content"},
		})
		respMsg := &importv1.ParseTransactionsResponse{
			Transactions: []*importv1.ParseTransactionsResponse_ParsedTransaction{
				{DuplicateTransactionId: nil},
			},
		}
		mockSvc.EXPECT().Parse(gomock.Any(), req.Msg).Return(respMsg, nil)
		resp, err := api.ParseTransactions(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
		assert.Len(t, resp.Msg.Transactions, 1)
	})

	t.Run("service error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&importv1.ParseTransactionsRequest{
			Source:  importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
			Content: []string{"test content"},
		})
		mockSvc.EXPECT().Parse(gomock.Any(), req.Msg).Return(nil, assert.AnError)
		resp, err := api.ParseTransactions(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&importv1.ParseTransactionsRequest{
			Source:  importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
			Content: []string{"test content"},
		})
		resp, err := api.ParseTransactions(context.TODO(), req)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})
}

func TestImportApi_MarkTransactionsIgnored_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockImportSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewImportApi(grpc, mockSvc)

	t.Run("marks references", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&importv1.MarkTransactionsIgnoredRequest{
			ImportSource:     importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
			ReferenceNumbers: []string{"a", "b", "c"},
			Reason:           lo.ToPtr("manual"),
		})
		respMsg := &importv1.MarkTransactionsIgnoredResponse{IgnoredCount: 3}
		mockSvc.EXPECT().MarkTransactionsIgnored(gomock.Any(), req.Msg).
			DoAndReturn(func(_ context.Context, r *importv1.MarkTransactionsIgnoredRequest) (*importv1.MarkTransactionsIgnoredResponse, error) {
				assert.Equal(t, importv1.ImportSource_IMPORT_SOURCE_FIREFLY, r.ImportSource)
				assert.Equal(t, []string{"a", "b", "c"}, r.ReferenceNumbers)
				assert.Equal(t, "manual", lo.FromPtr(r.Reason))
				return respMsg, nil
			})
		resp, err := api.MarkTransactionsIgnored(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, respMsg, resp.Msg)
		assert.EqualValues(t, 3, resp.Msg.IgnoredCount)
	})
}

func TestImportApi_MarkTransactionsIgnored_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockSvc := NewMockImportSvc(ctrl)
	grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
	api, _ := handlers.NewImportApi(grpc, mockSvc)

	t.Run("no auth", func(t *testing.T) {
		req := connect.NewRequest(&importv1.MarkTransactionsIgnoredRequest{
			ImportSource:     importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
			ReferenceNumbers: []string{"a"},
		})
		resp, err := api.MarkTransactionsIgnored(context.TODO(), req)
		assert.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})

	t.Run("invalid argument", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&importv1.MarkTransactionsIgnoredRequest{
			ImportSource:     importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
			ReferenceNumbers: []string{},
		})
		mockSvc.EXPECT().MarkTransactionsIgnored(gomock.Any(), req.Msg).
			Return(nil, importers.ErrNoReferenceNumbers)
		resp, err := api.MarkTransactionsIgnored(ctx, req)
		assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
		assert.ErrorIs(t, err, importers.ErrNoReferenceNumbers)
		assert.Nil(t, resp)
	})

	t.Run("internal error", func(t *testing.T) {
		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&importv1.MarkTransactionsIgnoredRequest{
			ImportSource:     importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
			ReferenceNumbers: []string{"a"},
		})
		mockSvc.EXPECT().MarkTransactionsIgnored(gomock.Any(), req.Msg).
			Return(nil, errors.New("boom"))
		resp, err := api.MarkTransactionsIgnored(ctx, req)
		assert.Equal(t, connect.CodeInternal, connect.CodeOf(err))
		assert.Nil(t, resp)
	})
}
