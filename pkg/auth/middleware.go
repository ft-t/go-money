package auth

import (
	"connectrpc.com/connect"
	"context"
)

var GrpcMiddleware = func(jwtParser JwtValidator) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			jwt := request.Header().Get("authorization")

			if len(jwt) == 0 {
				return next(ctx, request)
			}

			jwt = jwt[len("Bearer "):]

			parsed, err := jwtParser.ValidateToken(ctx, jwt)
			if err != nil {
				return nil, err
			}

			ctx = WithContext(ctx, *parsed)

			return next(ctx, request)
		}
	}
}

type jwtKey struct{}

func WithContext(ctx context.Context, wrapper JwtClaims) context.Context {
	return context.WithValue(ctx, jwtKey{}, &wrapper)
}

func FromContext(ctx context.Context) JwtClaims {
	val := ctx.Value(jwtKey{})
	if val == nil {
		return JwtClaims{}
	}

	return *val.(*JwtClaims)
}
