package main

import (
	"connectrpc.com/connect"
	"context"
	usersv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/users/v1"
	"github.com/ft-t/go-money-pb/gen/gomoneypb/users/v1/usersv1connect"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type UserApi struct {
}

func (u *UserApi) Login(
	ctx context.Context,
	c *connect.Request[usersv1.LoginRequest],
) (*connect.Response[usersv1.LoginResponse], error) {
	//TODO implement me
	panic("implement me")
}

func NewUserApi(
	mux *boilerplate.DefaultGrpcServer,
) (*UserApi, error) {
	res := &UserApi{}

	mux.GetMux().Handle(
		usersv1connect.NewUsersServiceHandler(res, mux.GetDefaultHandlerOptions()...),
	)

	return res, nil
}
