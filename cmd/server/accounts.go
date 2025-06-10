package main

import (
	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/accounts/v1/accountsv1connect"
	accountsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/accounts/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type AccountsApi struct {
	accSvc AccountSvc
}

func (a *AccountsApi) ReorderAccounts(ctx context.Context, c *connect.Request[accountsv1.ReorderAccountsRequest]) (*connect.Response[accountsv1.ReorderAccountsResponse], error) {

	//TODO implement me
	panic("implement me")
}

func (a *AccountsApi) DeleteAccount(ctx context.Context, c *connect.Request[accountsv1.DeleteAccountRequest]) (*connect.Response[accountsv1.DeleteAccountResponse], error) {
	// todo auth

	resp, err := a.accSvc.Delete(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (a *AccountsApi) CreateAccount(
	ctx context.Context,
	c *connect.Request[accountsv1.CreateAccountRequest],
) (*connect.Response[accountsv1.CreateAccountResponse], error) {
	// todo auth

	resp, err := a.accSvc.Create(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (a *AccountsApi) UpdateAccount(
	ctx context.Context,
	c *connect.Request[accountsv1.UpdateAccountRequest],
) (*connect.Response[accountsv1.UpdateAccountResponse], error) {
	// todo auth

	resp, err := a.accSvc.Update(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func (a *AccountsApi) ListAccounts(
	ctx context.Context,
	c *connect.Request[accountsv1.ListAccountsRequest],
) (*connect.Response[accountsv1.ListAccountsResponse], error) {
	// todo auth
	resp, err := a.accSvc.List(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func NewAccountsApi(
	mux *boilerplate.DefaultGrpcServer,
	accSvc AccountSvc,
) (*AccountsApi, error) {
	res := &AccountsApi{
		accSvc: accSvc,
	}

	mux.GetMux().Handle(
		accountsv1connect.NewAccountsServiceHandler(res, mux.GetDefaultHandlerOptions()...),
	)

	return res, nil
}
