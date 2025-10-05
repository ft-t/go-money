package handlers

import (
	"context"

	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/analytics/v1/analyticsv1connect"
	analyticsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/analytics/v1"
	"connectrpc.com/connect"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type AnalyticsApi struct {
	analyticsSvc AnalyticsSvc
}

func (a *AnalyticsApi) GetDebitsAndCreditsSummary(ctx context.Context, c *connect.Request[analyticsv1.GetDebitsAndCreditsSummaryRequest]) (*connect.Response[analyticsv1.GetDebitsAndCreditsSummaryResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := a.analyticsSvc.GetDebitsAndCreditsSummary(ctx, c.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(resp), nil
}

func NewAnalyticsApi(
	mux *boilerplate.DefaultGrpcServer,
	analyticsSvc AnalyticsSvc,
) *AnalyticsApi {
	res := &AnalyticsApi{
		analyticsSvc: analyticsSvc,
	}
	mux.GetMux().Handle(
		analyticsv1connect.NewAnalyticsServiceHandler(res, mux.GetDefaultHandlerOptions()...),
	)

	return res
}
