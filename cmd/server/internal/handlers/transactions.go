package handlers

import (
	"context"

	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/transactions/v1/transactionsv1connect"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	"connectrpc.com/connect"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type TransactionApi struct {
	transactionsSvc      TransactionsSvc
	applicableAccountSvc ApplicableAccountSvc
	mapper               MapperSvc
}

func (a *TransactionApi) CreateTransactionsBulk(ctx context.Context, c *connect.Request[transactionsv1.CreateTransactionsBulkRequest]) (*connect.Response[transactionsv1.CreateTransactionsBulkResponse], error) {
	//TODO implement me
	panic("implement me")
}

func (a *TransactionApi) DeleteTransactions(ctx context.Context, c *connect.Request[transactionsv1.DeleteTransactionsRequest]) (*connect.Response[transactionsv1.DeleteTransactionsRequest], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	_, err := a.transactionsSvc.DeleteTransaction(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(c.Msg), nil
}

func (a *TransactionApi) GetApplicableAccounts(
	ctx context.Context,
	_ *connect.Request[transactionsv1.GetApplicableAccountsRequest],
) (*connect.Response[transactionsv1.GetApplicableAccountsResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := a.applicableAccountSvc.GetAll(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := &transactionsv1.GetApplicableAccountsResponse{}
	for txType, account := range resp {
		rec := &transactionsv1.GetApplicableAccountsResponse_ApplicableRecord{
			TransactionType: txType,
		}

		for _, source := range account.SourceAccounts {
			rec.SourceAccounts = append(rec.SourceAccounts, a.mapper.MapAccount(ctx, source))
		}
		for _, dest := range account.DestinationAccounts {
			rec.DestinationAccounts = append(rec.DestinationAccounts, a.mapper.MapAccount(ctx, dest))
		}

		res.ApplicableRecords = append(res.ApplicableRecords, rec)
	}

	return connect.NewResponse(res), nil
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

func (a *TransactionApi) GetTitleSuggestions(ctx context.Context, c *connect.Request[transactionsv1.GetTitleSuggestionsRequest]) (*connect.Response[transactionsv1.GetTitleSuggestionsResponse], error) {
	jwtData := middlewares.FromContext(ctx)
	if jwtData.UserID == 0 {
		return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
	}

	resp, err := a.transactionsSvc.GetTitleSuggestions(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func NewTransactionApi(
	mux *boilerplate.DefaultGrpcServer,
	transactionsSvc TransactionsSvc,
	applicableAccountSvc ApplicableAccountSvc,
	mapper MapperSvc,
) *TransactionApi {
	res := &TransactionApi{
		transactionsSvc:      transactionsSvc,
		applicableAccountSvc: applicableAccountSvc,
		mapper:               mapper,
	}

	mux.GetMux().Handle(
		transactionsv1connect.NewTransactionsServiceHandler(res, mux.GetDefaultHandlerOptions()...))

	return res
}
