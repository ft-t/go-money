package middlewares

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/ft-t/go-money/pkg/auth"
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
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
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

func HTTPAuthMiddleware(jwtParser JwtValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := authHeader[7:]

		claims, err := jwtParser.ValidateToken(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := WithContext(r.Context(), *claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
