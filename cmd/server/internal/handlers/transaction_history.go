package handlers

import (
	"context"

	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/transactions/history/v1/historyv1connect"
	historyv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/history/v1"
	"connectrpc.com/connect"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type TransactionHistoryApi struct {
	svc    TransactionHistorySvc
	mapper TransactionHistoryMapper
}

func NewTransactionHistoryApi(
	mux *boilerplate.DefaultGrpcServer,
	svc TransactionHistorySvc,
	mapper TransactionHistoryMapper,
) (*TransactionHistoryApi, error) {
	res := &TransactionHistoryApi{svc: svc, mapper: mapper}

	mux.GetMux().Handle(
		historyv1connect.NewTransactionHistoryServiceHandler(res, mux.GetDefaultHandlerOptions()...),
	)

	return res, nil
}

func (a *TransactionHistoryApi) ListHistory(
	ctx context.Context,
	c *connect.Request[historyv1.ListTransactionHistoryRequest],
) (*connect.Response[historyv1.ListTransactionHistoryResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	rows, err := a.svc.List(ctx, c.Msg.TransactionId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	out := make([]*historyv1.TransactionHistoryEvent, 0, len(rows))
	for _, r := range rows {
		out = append(out, a.mapper.MapTransactionHistoryEvent(r))
	}

	return connect.NewResponse(&historyv1.ListTransactionHistoryResponse{Events: out}), nil
}
