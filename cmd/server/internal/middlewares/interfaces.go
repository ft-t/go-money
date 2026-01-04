package middlewares

import (
	"context"

	"github.com/ft-t/go-money/pkg/auth"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package middlewares_test -source=interfaces.go

type JwtValidator interface {
	ValidateToken(
		_ context.Context,
		tokenString string,
	) (*auth.JwtClaims, error)
}

type ServiceTokenValidator interface {
	IsRevoked(ctx context.Context, tokenID string) (bool, error)
}
