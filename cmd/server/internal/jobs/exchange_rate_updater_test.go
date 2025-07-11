package jobs_test

import (
	"context"
	"github.com/ft-t/go-money/cmd/server/internal/jobs"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUpdateCurrencyRates(t *testing.T) {
	sync := NewMockCurrencySyncerSvc(gomock.NewController(t))

	scheduler, err := jobs.NewJobScheduler(&jobs.Config{
		ExchangeRatesUpdateSvc: sync,
		Configuration: configuration.Configuration{
			ExchangeRatesUrl: "https://test.com",
		},
	})
	assert.NoError(t, err)

	sync.EXPECT().Sync(gomock.Any(), "https://test.com").Return(nil)

	assert.NoError(t, scheduler.UpdateCurrencyRates(context.TODO()))
}
