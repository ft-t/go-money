package auth

import "context"

type JwtSvc interface {
	CreateServiceToken(
		_ context.Context,
		req *GenerateTokenRequest,
	) (*JwtClaims, string, error)
}
