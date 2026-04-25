package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	historyv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/history/v1"
	"connectrpc.com/connect"
	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionHistoryListHistory_Success(t *testing.T) {
	t.Run("returns mapped events", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		svc := NewMockTransactionHistorySvc(ctrl)
		mapper := NewMockTransactionHistoryMapper(ctrl)
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()

		app, err := handlers.NewTransactionHistoryApi(grpc, svc, mapper)
		require.NoError(t, err)
		require.NotNil(t, app)

		row1 := &database.TransactionHistory{ID: 1, TransactionID: 10}
		row2 := &database.TransactionHistory{ID: 2, TransactionID: 10}
		ev1 := &historyv1.TransactionHistoryEvent{Id: 1, TransactionId: 10}
		ev2 := &historyv1.TransactionHistoryEvent{Id: 2, TransactionId: 10}

		svc.EXPECT().List(gomock.Any(), int64(10)).Return([]*database.TransactionHistory{row1, row2}, nil)
		mapper.EXPECT().MapTransactionHistoryEvent(row1).Return(ev1)
		mapper.EXPECT().MapTransactionHistoryEvent(row2).Return(ev2)

		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 123})
		req := connect.NewRequest(&historyv1.ListTransactionHistoryRequest{TransactionId: 10})

		resp, err := app.ListHistory(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, []*historyv1.TransactionHistoryEvent{ev1, ev2}, resp.Msg.Events)
	})

	t.Run("empty result", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		svc := NewMockTransactionHistorySvc(ctrl)
		mapper := NewMockTransactionHistoryMapper(ctrl)
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()

		app, err := handlers.NewTransactionHistoryApi(grpc, svc, mapper)
		require.NoError(t, err)

		svc.EXPECT().List(gomock.Any(), int64(99)).Return([]*database.TransactionHistory{}, nil)

		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})
		req := connect.NewRequest(&historyv1.ListTransactionHistoryRequest{TransactionId: 99})

		resp, err := app.ListHistory(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Empty(t, resp.Msg.Events)
	})
}

func TestTransactionHistoryListHistory_Failure(t *testing.T) {
	t.Run("no auth", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		svc := NewMockTransactionHistorySvc(ctrl)
		mapper := NewMockTransactionHistoryMapper(ctrl)
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()

		app, err := handlers.NewTransactionHistoryApi(grpc, svc, mapper)
		require.NoError(t, err)

		req := connect.NewRequest(&historyv1.ListTransactionHistoryRequest{TransactionId: 10})
		resp, err := app.ListHistory(context.TODO(), req)

		assert.ErrorIs(t, err, auth.ErrInvalidToken)
		assert.Nil(t, resp)
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		svc := NewMockTransactionHistorySvc(ctrl)
		mapper := NewMockTransactionHistoryMapper(ctrl)
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()

		app, err := handlers.NewTransactionHistoryApi(grpc, svc, mapper)
		require.NoError(t, err)

		svcErr := errors.New("boom")
		svc.EXPECT().List(gomock.Any(), int64(10)).Return(nil, svcErr)

		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 123})
		req := connect.NewRequest(&historyv1.ListTransactionHistoryRequest{TransactionId: 10})

		resp, err := app.ListHistory(ctx, req)

		assert.ErrorIs(t, err, svcErr)
		assert.Nil(t, resp)
	})
}
