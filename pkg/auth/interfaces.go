package auth

import (
	"context"
)

type JwtValidator interface {
	ValidateToken(
		_ context.Context,
		tokenString string,
	) (*JwtClaims, error)
}
