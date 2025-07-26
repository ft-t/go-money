package rules_test

import (
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/go-co-op/gocron/v2"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewScheduler(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		sh := rules.NewScheduler(&rules.SchedulerConfig{})
		assert.NoError(t, sh.ValidateCronExpression("*/5 * * * *"))

		assert.NoError(t, sh.SchedulerTestTask())
	})

	t.Run("invalid cron expression", func(t *testing.T) {
		sh := rules.NewScheduler(&rules.SchedulerConfig{})
		assert.ErrorContains(t, sh.ValidateCronExpression("*/x * * * *"), "crontab parse failure")
	})

	t.Run("fail", func(t *testing.T) {
		sh := rules.NewScheduler(&rules.SchedulerConfig{
			Opts: []gocron.SchedulerOption{
				gocron.WithDistributedLocker(nil),
			},
		})

		assert.ErrorContains(t, sh.ValidateCronExpression("*/5 * * * *"), "locker must not be nil")
	})
}
