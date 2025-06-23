package middlewares_test

import (
	"connectrpc.com/connect"
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGrpcMiddleware(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx := context.TODO()

		parser := NewMockJwtValidator(gomock.NewController(t))
		req := connect.NewRequest[any](nil)
		req.Header().Set("Authorization", "Bearer valid_token")

		jwtData := auth.JwtClaims{
			UserID: 123,
		}

		parser.EXPECT().ValidateToken(gomock.Any(), "valid_token").Return(&jwtData, nil)

		called := false

		response, err := middlewares.GrpcMiddleware(parser)(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			assert.EqualValues(t, jwtData, middlewares.FromContext(ctx))
			called = true

			return &connect.Response[any]{}, nil
		})(ctx, req)

		assert.True(t, called)
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})

	t.Run("invalid token", func(t *testing.T) {
		ctx := context.TODO()

		parser := NewMockJwtValidator(gomock.NewController(t))
		req := connect.NewRequest[any](nil)
		req.Header().Set("Authorization", "Bearer valid_token")

		parser.EXPECT().ValidateToken(gomock.Any(), "valid_token").Return(nil, errors.New("invalid token"))

		called := false
		response, err := middlewares.GrpcMiddleware(parser)(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			called = true
			return &connect.Response[any]{}, nil
		})(ctx, req)

		assert.False(t, called)
		assert.ErrorContains(t, err, "invalid token")
		assert.Nil(t, response)
	})

	t.Run("no token", func(t *testing.T) {
		ctx := context.TODO()

		called := false

		parser := NewMockJwtValidator(gomock.NewController(t))
		response, err := middlewares.GrpcMiddleware(parser)(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			jwtData := middlewares.FromContext(ctx)
			assert.EqualValues(t, 0, jwtData.UserID)
			called = true
			return &connect.Response[any]{}, nil
		})(ctx, &connect.Request[any]{})

		assert.True(t, called)
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})
}
