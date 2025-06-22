package auth

import (
	"context"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package auth_test -source=interfaces.go

type JwtValidator interface {
	ValidateToken(
		_ context.Context,
		tokenString string,
	) (*JwtClaims, error)
}
