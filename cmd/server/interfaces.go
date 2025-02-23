package main

import (
	"context"
	configurationv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/configuration/v1"
	usersv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/users/v1"
)

type UserSvc interface {
	Login(
		ctx context.Context,
		req *usersv1.LoginRequest,
	) (*usersv1.LoginResponse, error)

	Create(
		ctx context.Context,
		req *usersv1.CreateRequest,
	) (*usersv1.CreateResponse, error)
}

type ConfigSvc interface {
	GetConfiguration(
		ctx context.Context,
		_ *configurationv1.GetConfigurationRequest,
	) (*configurationv1.GetConfigurationResponse, error)
}
