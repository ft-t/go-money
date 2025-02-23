package main

import (
	"context"
	usersv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/users/v1"
)

type UserSvc interface {
	Login(
		ctx context.Context,
		req *usersv1.LoginRequest,
	) (*usersv1.LoginResponse, error)
}
