package handlers

import (
	"context"

	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/maintenance/v1/maintenancev1connect"
	maintenancev1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/maintenance/v1"
	"connectrpc.com/connect"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type MaintenanceApi struct {
	recalculateSvc RecalculateSvc
}

func (m *MaintenanceApi) RecalculateAll(
	ctx context.Context,
	_ *connect.Request[maintenancev1.RecalculateAllRequest],
) (*connect.Response[maintenancev1.RecalculateAllResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	if err := m.recalculateSvc.RecalculateAll(ctx); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&maintenancev1.RecalculateAllResponse{
		Success: true,
	}), nil
}

func NewMaintenanceApi(
	mux *boilerplate.DefaultGrpcServer,
	recalculateSvc RecalculateSvc,
) *MaintenanceApi {
	res := &MaintenanceApi{
		recalculateSvc: recalculateSvc,
	}

	mux.GetMux().Handle(
		maintenancev1connect.NewMaintenanceServiceHandler(res, mux.GetDefaultHandlerOptions()...),
	)

	return res
}
