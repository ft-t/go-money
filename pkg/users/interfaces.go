package users

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
)

type JwtSvc interface {
	GenerateToken(
		_ context.Context,
		user *database.User,
	) (string, error)
}
