package auth

import (
	"context"
	"time"
)

type JwtSvc interface {
	CreateServiceToken(
		_ context.Context,
		req *GenerateTokenRequest,
	) (*JwtClaims, string, error)

	RevokeServiceToken(
		ctx context.Context,
		jti string,
		originalExpiresAt time.Time,
	) error
}
