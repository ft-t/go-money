package configuration

import "github.com/ft-t/go-money/pkg/boilerplate"

type Configuration struct {
	Db            boilerplate.DbConfig `env:", prefix=DB_"`
	ReadOnlyDb    boilerplate.DbConfig `env:", prefix=READONLY_DB_"`
	GrpcPort      int                  `env:"GRPC_PORT"`
	JwtPrivateKey string               `env:"JWT_PRIVATE_KEY"`
}
