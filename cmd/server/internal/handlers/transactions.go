package handlers

import (
	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/transactions/v1/transactionsv1connect"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type TransactionApi struct {
	transactionsSvc      TransactionsSvc
	applicableAccountSvc ApplicableAccountSvc
}

func (a *TransactionApi) GetApplicableAccounts(ctx context.Context, c *connect.Request[transactionsv1.GetApplicableAccountsRequest]) (*connect.Response[transactionsv1.GetApplicableAccountsResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	// todo
}

func (a *TransactionApi) ListTransactions(ctx context.Context, c *connect.Request[transactionsv1.ListTransactionsRequest]) (*connect.Response[transactionsv1.ListTransactionsResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := a.transactionsSvc.List(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (a *TransactionApi) CreateTransaction(
	ctx context.Context,
	c *connect.Request[transactionsv1.CreateTransactionRequest],
) (*connect.Response[transactionsv1.CreateTransactionResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := a.transactionsSvc.Create(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (a *TransactionApi) UpdateTransaction(ctx context.Context, c *connect.Request[transactionsv1.UpdateTransactionRequest]) (*connect.Response[transactionsv1.UpdateTransactionResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := a.transactionsSvc.Update(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func NewTransactionApi(
	mux *boilerplate.DefaultGrpcServer,
	transactionsSvc TransactionsSvc,
	applicableAccountSvc ApplicableAccountSvc,
) *TransactionApi {
	res := &TransactionApi{
		transactionsSvc:      transactionsSvc,
		applicableAccountSvc: applicableAccountSvc,
	}

	mux.GetMux().Handle(
		transactionsv1connect.NewTransactionsServiceHandler(res, mux.GetDefaultHandlerOptions()...))

	return res
}
