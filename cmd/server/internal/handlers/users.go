package handlers

import (
	"buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/users/v1/usersv1connect"
	usersv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/users/v1"
	"connectrpc.com/connect"
	"context"
	"github.com/ft-t/go-money/pkg/boilerplate"
)

type UserApi struct {
	userSvc UserSvc
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

func (u *UserApi) Create(ctx context.Context, c *connect.Request[usersv1.CreateRequest]) (*connect.Response[usersv1.CreateResponse], error) {
	resp, err := u.userSvc.Create(ctx, c.Msg) // check is basically inside
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(resp), nil
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
