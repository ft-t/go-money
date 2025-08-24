package handlers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/ft-t/go-money/cmd/server/internal/middlewares"
	"github.com/ft-t/go-money/pkg/auth"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRecalculateAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		recalculateSvc := NewMockRecalculateSvc(gomock.NewController(t))

		api := handlers.NewMaintenanceApi(grpc, recalculateSvc)

		recalculateSvc.EXPECT().RecalculateAll(gomock.All()).
			Return(nil)

		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})

		resp, err := api.RecalculateAll(ctx, nil)
		assert.NoError(t, err)
		assert.True(t, resp.Msg.Success)
	})

	t.Run("recalculate error", func(t *testing.T) {
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		recalculateSvc := NewMockRecalculateSvc(gomock.NewController(t))

		api := handlers.NewMaintenanceApi(grpc, recalculateSvc)

		recalculateSvc.EXPECT().RecalculateAll(gomock.All()).
			Return(assert.AnError)

		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 1})

		_, err := api.RecalculateAll(ctx, nil)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("no user id", func(t *testing.T) {
		grpc := boilerplate.NewDefaultGrpcServerBuild(http.NewServeMux()).Build()
		recalculateSvc := NewMockRecalculateSvc(gomock.NewController(t))

		api := handlers.NewMaintenanceApi(grpc, recalculateSvc)

		ctx := middlewares.WithContext(context.TODO(), auth.JwtClaims{UserID: 0})

		_, err := api.RecalculateAll(ctx, nil)
		assert.ErrorIs(t, err, auth.ErrInvalidToken)
	})
}
