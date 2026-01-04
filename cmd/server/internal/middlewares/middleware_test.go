package middlewares_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/cockroachdb/errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
)

func TestGrpcMiddleware_Success(t *testing.T) {
	t.Run("web token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ctx := context.TODO()

		parser := NewMockJwtValidator(ctrl)
		serviceTokenValidator := NewMockServiceTokenValidator(ctrl)
		req := connect.NewRequest[any](nil)
		req.Header().Set("Authorization", "Bearer valid_token")

		jwtData := auth.JwtClaims{
			UserID:    123,
			TokenType: "web",
		}

		parser.EXPECT().ValidateToken(gomock.Any(), "valid_token").Return(&jwtData, nil)

		called := false

		response, err := middlewares.GrpcMiddleware(parser, serviceTokenValidator)(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			assert.EqualValues(t, jwtData, middlewares.FromContext(ctx))
			called = true

			return &connect.Response[any]{}, nil
		})(ctx, req)

		assert.True(t, called)
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})

	t.Run("service token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ctx := context.TODO()

		parser := NewMockJwtValidator(ctrl)
		serviceTokenValidator := NewMockServiceTokenValidator(ctrl)
		req := connect.NewRequest[any](nil)
		req.Header().Set("Authorization", "Bearer valid_token")

		jwtData := auth.JwtClaims{
			RegisteredClaims: &jwt.RegisteredClaims{
				ID: "token-id-123",
			},
			UserID:    123,
			TokenType: auth.ServiceTokenType,
		}

		parser.EXPECT().ValidateToken(gomock.Any(), "valid_token").Return(&jwtData, nil)
		serviceTokenValidator.EXPECT().IsRevoked(gomock.Any(), "token-id-123").Return(false, nil)

		called := false

		response, err := middlewares.GrpcMiddleware(parser, serviceTokenValidator)(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			assert.EqualValues(t, jwtData, middlewares.FromContext(ctx))
			called = true

			return &connect.Response[any]{}, nil
		})(ctx, req)

		assert.True(t, called)
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})
}

func TestGrpcMiddleware_Failure(t *testing.T) {
	t.Run("invalid token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ctx := context.TODO()

		parser := NewMockJwtValidator(ctrl)
		serviceTokenValidator := NewMockServiceTokenValidator(ctrl)
		req := connect.NewRequest[any](nil)
		req.Header().Set("Authorization", "Bearer valid_token")

		parser.EXPECT().ValidateToken(gomock.Any(), "valid_token").Return(nil, errors.New("invalid token"))

		called := false
		response, err := middlewares.GrpcMiddleware(parser, serviceTokenValidator)(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			called = true
			return &connect.Response[any]{}, nil
		})(ctx, req)

		assert.False(t, called)
		assert.ErrorContains(t, err, "invalid token")
		assert.Nil(t, response)

		var connectErr *connect.Error
		if errors.As(err, &connectErr) {
			assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code())
		} else {
			t.Fatalf("expected connect.Error, got %T", err)
		}
	})

	t.Run("service token revoked", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ctx := context.TODO()

		parser := NewMockJwtValidator(ctrl)
		serviceTokenValidator := NewMockServiceTokenValidator(ctrl)
		req := connect.NewRequest[any](nil)
		req.Header().Set("Authorization", "Bearer valid_token")

		jwtData := auth.JwtClaims{
			RegisteredClaims: &jwt.RegisteredClaims{
				ID: "revoked-token-id",
			},
			UserID:    123,
			TokenType: auth.ServiceTokenType,
		}

		parser.EXPECT().ValidateToken(gomock.Any(), "valid_token").Return(&jwtData, nil)
		serviceTokenValidator.EXPECT().IsRevoked(gomock.Any(), "revoked-token-id").Return(true, nil)

		called := false
		response, err := middlewares.GrpcMiddleware(parser, serviceTokenValidator)(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
			called = true
			return &connect.Response[any]{}, nil
		})(ctx, req)

		assert.False(t, called)
		assert.ErrorIs(t, err, middlewares.ErrTokenRevoked)
		assert.Nil(t, response)
	})
}

func TestGrpcMiddleware_NoToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.TODO()

	called := false

	parser := NewMockJwtValidator(ctrl)
	serviceTokenValidator := NewMockServiceTokenValidator(ctrl)
	response, err := middlewares.GrpcMiddleware(parser, serviceTokenValidator)(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		jwtData := middlewares.FromContext(ctx)
		assert.EqualValues(t, 0, jwtData.UserID)
		called = true
		return &connect.Response[any]{}, nil
	})(ctx, &connect.Request[any]{})

	assert.True(t, called)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}
