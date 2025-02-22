package configuration

import "go-money/pkg/boilerplate"

type Configuration struct {
	Db         boilerplate.DbConfig
	ReadOnlyDb boilerplate.DbConfig
}
