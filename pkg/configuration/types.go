package configuration

import "github.com/ft-t/go-money/pkg/boilerplate"

type Configuration struct {
	Db            boilerplate.DbConfig
	ReadOnlyDb    boilerplate.DbConfig
	GrpcPort      int
	JwtPrivateKey string
}
