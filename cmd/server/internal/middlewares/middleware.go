package middlewares

import (
	"context"

	"connectrpc.com/connect"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/auth"
)

var ErrTokenRevoked = errors.New("token has been revoked")

var GrpcMiddleware = func(jwtParser JwtValidator, serviceTokenValidator ServiceTokenValidator) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			jwt := request.Header().Get("authorization")

			if len(jwt) == 0 {
				return next(ctx, request)
			}

			jwt = jwt[len("Bearer "):]

			parsed, err := jwtParser.ValidateToken(ctx, jwt)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			if parsed.TokenType == auth.ServiceTokenType {
				revoked, revokeErr := serviceTokenValidator.IsRevoked(ctx, parsed.ID)
				if revokeErr != nil {
					return nil, connect.NewError(connect.CodeInternal, revokeErr)
				}
				if revoked {
					return nil, connect.NewError(connect.CodeUnauthenticated, ErrTokenRevoked)
				}
			}

			ctx = WithContext(ctx, *parsed)

			return next(ctx, request)
		}
	}
}

type jwtKey struct{}

func WithContext(ctx context.Context, wrapper auth.JwtClaims) context.Context {
	return context.WithValue(ctx, jwtKey{}, &wrapper)
}

func FromContext(ctx context.Context) auth.JwtClaims {
	val := ctx.Value(jwtKey{})
	if val == nil {
		return auth.JwtClaims{}
	}

	return *val.(*auth.JwtClaims)
}
