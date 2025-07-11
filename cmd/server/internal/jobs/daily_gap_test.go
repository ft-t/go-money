package jobs_test

import (
	"context"
	"github.com/ft-t/go-money/cmd/server/internal/jobs"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFixDailyGaps(t *testing.T) {
	sync := NewMockMaintenanceSvc(gomock.NewController(t))

	scheduler, err := jobs.NewJobScheduler(&jobs.Config{
		MaintenanceSvc: sync,
		Configuration:  configuration.Configuration{},
	})
	assert.NoError(t, err)

	sync.EXPECT().FixDailyGaps(gomock.Any()).Return(nil)

	assert.NoError(t, scheduler.FixDailyGap(context.TODO()))
}
