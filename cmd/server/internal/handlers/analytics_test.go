package handlers_test

import (
	"context"
	"net/http"
	"testing"

	analyticsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/analytics/v1"
	"connectrpc.com/connect"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAnalyticsApi_GetDebitsAndCreditsSummary(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		analyticsSvc := NewMockAnalyticsSvc(ctrl)
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()

		api := handlers.NewAnalyticsApi(grpc, analyticsSvc)
		assert.NotNil(t, api)

		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{
			UserID: 123,
		})

		req := &analyticsv1.GetDebitsAndCreditsSummaryRequest{
			AccountIds: []int32{int32(123)},
			StartAt:    timestamppb.Now(),
			EndAt:      timestamppb.Now(),
		}

		expectedResp := &analyticsv1.GetDebitsAndCreditsSummaryResponse{
			Items: map[int32]*analyticsv1.GetDebitsAndCreditsSummaryResponse_SummaryItem{
				123: {
					TotalDebitsCount:   2,
					TotalCreditsCount:  3,
					TotalDebitsAmount:  "1000.00",
					TotalCreditsAmount: "2000.00",
				},
			},
		}

		analyticsSvc.EXPECT().
			GetDebitsAndCreditsSummary(ctx, req).
			Return(expectedResp, nil).
			Times(1)

		connectReq := connect.NewRequest(req)
		resp, err := api.GetDebitsAndCreditsSummary(ctx, connectReq)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedResp, resp.Msg)
	})

	t.Run("service error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		analyticsSvc := NewMockAnalyticsSvc(ctrl)
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()

		api := handlers.NewAnalyticsApi(grpc, analyticsSvc)
		assert.NotNil(t, api)

		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{
			UserID: 123,
		})

		req := &analyticsv1.GetDebitsAndCreditsSummaryRequest{
			AccountIds: []int32{int32(123)},
		}

		serviceError := errors.New("service error")
		analyticsSvc.EXPECT().
			GetDebitsAndCreditsSummary(ctx, req).
			Return(nil, serviceError).
			Times(1)

		connectReq := connect.NewRequest(req)
		resp, err := api.GetDebitsAndCreditsSummary(ctx, connectReq)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "service error")

		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodeInternal, connectErr.Code())
	})

	t.Run("no authentication", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		analyticsSvc := NewMockAnalyticsSvc(ctrl)
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()

		api := handlers.NewAnalyticsApi(grpc, analyticsSvc)
		assert.NotNil(t, api)

		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{
			UserID: 0,
		})

		req := &analyticsv1.GetDebitsAndCreditsSummaryRequest{
			AccountIds: []int32{int32(123)},
		}

		analyticsSvc.EXPECT().
			GetDebitsAndCreditsSummary(gomock.Any(), gomock.Any()).
			Times(0)

		connectReq := connect.NewRequest(req)
		resp, err := api.GetDebitsAndCreditsSummary(ctx, connectReq)

		assert.Error(t, err)
		assert.Nil(t, resp)

		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodePermissionDenied, connectErr.Code())
		assert.ErrorIs(t, connectErr.Unwrap(), auth.ErrInvalidToken)
	})

	t.Run("empty account ids", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		analyticsSvc := NewMockAnalyticsSvc(ctrl)
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()

		api := handlers.NewAnalyticsApi(grpc, analyticsSvc)
		assert.NotNil(t, api)

		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{
			UserID: 123,
		})

		req := &analyticsv1.GetDebitsAndCreditsSummaryRequest{
			AccountIds: []int32{},
		}

		serviceError := errors.New("account_ids cannot be empty")
		analyticsSvc.EXPECT().
			GetDebitsAndCreditsSummary(ctx, req).
			Return(nil, serviceError).
			Times(1)

		connectReq := connect.NewRequest(req)
		resp, err := api.GetDebitsAndCreditsSummary(ctx, connectReq)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "account_ids cannot be empty")

		connectErr, ok := err.(*connect.Error)
		assert.True(t, ok)
		assert.Equal(t, connect.CodeInternal, connectErr.Code())
	})
}
