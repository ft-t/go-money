package appcfg_test

import (
	"context"
	"github.com/cockroachdb/errors"
	configurationv1 "github.com/ft-t/go-money-pb/gen/gomoneypb/configuration/v1"
	"github.com/ft-t/go-money/pkg/appcfg"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConfiguration(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		userSvc := NewMockUserSvc(gomock.NewController(t))
		userSvc.EXPECT().ShouldCreateAdmin(gomock.Any()).
			Return(true, nil)

		srv := appcfg.NewService(&appcfg.ServiceConfig{
			UserSvc: userSvc,
		})

		resp, err := srv.GetConfiguration(context.TODO(), &configurationv1.GetConfigurationRequest{})
		assert.NoError(t, err)
		assert.True(t, resp.ShouldCreateAdmin)
		assert.Equal(t, "USD", resp.BaseCurrency)
		assert.Equal(t, configuration.BaseCurrency, resp.BaseCurrency)
	})

	t.Run("fail", func(t *testing.T) {
		userSvc := NewMockUserSvc(gomock.NewController(t))
		userSvc.EXPECT().ShouldCreateAdmin(gomock.Any()).
			Return(false, errors.New("some error"))

		srv := appcfg.NewService(&appcfg.ServiceConfig{
			UserSvc: userSvc,
		})

		resp, err := srv.GetConfiguration(context.TODO(), &configurationv1.GetConfigurationRequest{})
		assert.ErrorContains(t, err, "some error")
		assert.Nil(t, resp)
	})
}
