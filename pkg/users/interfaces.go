package users

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package users_test -source=interfaces.go

type JwtSvc interface {
	GenerateToken(
		_ context.Context,
		user *database.User,
	) (string, error)
}
