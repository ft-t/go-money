package main

import (
	"connectrpc.com/connect"
	"context"
	usersv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/users/v1"
	"github.com/ft-t/go-money-pb/gen/gomoneypb/users/v1/usersv1connect"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type UserApi struct {
	userSvc UserSvc
}

func (u *UserApi) Login(
	ctx context.Context,
	c *connect.Request[usersv1.LoginRequest],
) (*connect.Response[usersv1.LoginResponse], error) {
	resp, err := u.userSvc.Login(ctx, c.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
}

func NewUserApi(
	mux *boilerplate.DefaultGrpcServer,
	userSvc UserSvc,
) (*UserApi, error) {
	res := &UserApi{
		userSvc: userSvc,
	}

	mux.GetMux().Handle(
		usersv1connect.NewUsersServiceHandler(res, mux.GetDefaultHandlerOptions()...),
	)

	return res, nil
}
