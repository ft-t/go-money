package main

import (
	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/transactions/v1/transactionsv1connect"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type TransactionApi struct {
	transactionsSvc TransactionsSvc
}

func (a *TransactionApi) ListTransactions(ctx context.Context, c *connect.Request[transactionsv1.ListTransactionsRequest]) (*connect.Response[transactionsv1.ListTransactionsResponse], error) {
	// todo auth

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
	// todo auth

	resp, err := a.transactionsSvc.Create(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func NewTransactionApi(
	mux *boilerplate.DefaultGrpcServer,
	transactionsSvc TransactionsSvc,
) *TransactionApi {
	res := &TransactionApi{
		transactionsSvc: transactionsSvc,
	}

	mux.GetMux().Handle(
		transactionsv1connect.NewTransactionsServiceHandler(res, mux.GetDefaultHandlerOptions()...))

	return res
}
