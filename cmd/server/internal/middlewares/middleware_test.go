package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/cockroachdb/errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
)

func TestGrpcMiddleware_Success(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ctx := context.TODO()

		parser := NewMockJwtValidator(ctrl)
		req := connect.NewRequest[any](nil)
		req.Header().Set("Authorization", "Bearer valid_token")

		jwtData := auth.JwtClaims{
			UserID:    123,
			TokenType: "web",
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
}

func TestGrpcMiddleware_Failure(t *testing.T) {
	t.Run("invalid token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		ctx := context.TODO()

		parser := NewMockJwtValidator(ctrl)
		req := connect.NewRequest[any](nil)
		req.Header().Set("Authorization", "Bearer invalid_token")

		parser.EXPECT().ValidateToken(gomock.Any(), "invalid_token").Return(nil, errors.New("invalid token"))

		called := false
		response, err := middlewares.GrpcMiddleware(parser)(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
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
}

func TestGrpcMiddleware_NoToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.TODO()

	called := false

	parser := NewMockJwtValidator(ctrl)
	response, err := middlewares.GrpcMiddleware(parser)(func(ctx context.Context, request connect.AnyRequest) (connect.AnyResponse, error) {
		jwtData := middlewares.FromContext(ctx)
		assert.EqualValues(t, 0, jwtData.UserID)
		called = true
		return &connect.Response[any]{}, nil
	})(ctx, &connect.Request[any]{})

	assert.True(t, called)
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestHTTPAuthMiddleware_Success(t *testing.T) {
	type tc struct {
		name   string
		token  string
		claims auth.JwtClaims
	}

	cases := []tc{
		{
			name:  "valid web token",
			token: "valid_web_token",
			claims: auth.JwtClaims{
				UserID:    123,
				TokenType: "web",
			},
		},
		{
			name:  "valid service token",
			token: "valid_service_token",
			claims: auth.JwtClaims{
				UserID:    456,
				TokenType: "service",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			parser := NewMockJwtValidator(ctrl)

			parser.EXPECT().ValidateToken(gomock.Any(), c.token).Return(&c.claims, nil)

			called := false
			handler := middlewares.HTTPAuthMiddleware(parser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims := middlewares.FromContext(r.Context())
				assert.Equal(t, c.claims.UserID, claims.UserID)
				assert.Equal(t, c.claims.TokenType, claims.TokenType)
				called = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+c.token)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.True(t, called)
			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestHTTPAuthMiddleware_Failure(t *testing.T) {
	type tc struct {
		name           string
		authHeader     string
		setupMock      func(*MockJwtValidator)
		expectedStatus int
		expectedBody   string
	}

	cases := []tc{
		{
			name:           "missing authorization header",
			authHeader:     "",
			setupMock:      func(m *MockJwtValidator) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "missing authorization header",
		},
		{
			name:           "invalid header format - no Bearer",
			authHeader:     "Basic token123",
			setupMock:      func(m *MockJwtValidator) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid authorization header format",
		},
		{
			name:           "invalid header format - too short",
			authHeader:     "Bear",
			setupMock:      func(m *MockJwtValidator) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid authorization header format",
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid_token",
			setupMock: func(m *MockJwtValidator) {
				m.EXPECT().ValidateToken(gomock.Any(), "invalid_token").Return(nil, errors.New("token expired"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid token",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			parser := NewMockJwtValidator(ctrl)
			c.setupMock(parser)

			called := false
			handler := middlewares.HTTPAuthMiddleware(parser, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if c.authHeader != "" {
				req.Header.Set("Authorization", c.authHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.False(t, called)
			assert.Equal(t, c.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), c.expectedBody)
		})
	}
}
