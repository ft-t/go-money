package jobs_test

import (
	"github.com/ft-t/go-money/cmd/server/internal/jobs"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/go-co-op/gocron/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJobScheduler(t *testing.T) {
	t.Run("succces", func(t *testing.T) {
		jobScheduler, err := jobs.NewJobScheduler(&jobs.Config{
			Configuration: configuration.Configuration{},
		})
		assert.NoError(t, err)
		assert.NotNil(t, jobScheduler)

		assert.NotNil(t, jobScheduler.GetScheduler())

		assert.NoError(t, jobScheduler.StartAsync())
		assert.NoError(t, jobScheduler.Stop())
	})

	t.Run("fail", func(t *testing.T) {
		jobScheduler, err := jobs.NewJobScheduler(&jobs.Config{
			Configuration: configuration.Configuration{},
			Opts: []gocron.SchedulerOption{
				gocron.WithDistributedLocker(nil),
			},
		})

		assert.ErrorContains(t, err, "locker must not be nil")
		assert.Nil(t, jobScheduler)
	})
}

//func TestJobScheduler_CurrencyRateUpdater(t *testing.T) {
//	jobScheduler, err := jobs.NewJobScheduler(&jobs.Config{
//		Configuration: configuration.Configuration{},
//	})
//	assert.NoError(t, err)
//	assert.NotNil(t, jobScheduler)
//
//	_ = jobScheduler.StartAsync()
//	time.Sleep(100 * time.Millisecond)
//
//
//	assert.NoError(t, jobScheduler.GetScheduler().Jobs()[0].RunNow())
//	time.Sleep(100 * time.Second)
//}
