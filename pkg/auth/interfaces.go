package auth

import (
	"context"
	"time"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"

	"github.com/ft-t/go-money/pkg/database"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package auth_test -source=interfaces.go

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

type ServiceTokenMapper interface {
	MapServiceToken(ctx context.Context, token *database.ServiceToken) *gomoneypbv1.ServiceToken
}
