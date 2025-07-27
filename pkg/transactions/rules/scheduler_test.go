package rules_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions/rules"
	"github.com/go-co-op/gocron/v2"
	"github.com/golang/mock/gomock"
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

	t.Run("fail cron", func(t *testing.T) {
		sh := rules.NewScheduler(&rules.SchedulerConfig{
			CronValidationOpts: []gocron.SchedulerOption{
				gocron.WithDistributedLocker(nil),
			},
		})

		assert.ErrorContains(t, sh.ValidateCronExpression("*/5 * * * *"), "locker must not be nil")
	})
}

func TestExecuteTask(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ruleInt := NewMockInterpreter(gomock.NewController(t))
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		sh := rules.NewScheduler(&rules.SchedulerConfig{
			RuleInterpreter: ruleInt,
			TransactionSvc:  txSvc,
		})

		ruleInt.EXPECT().Run(gomock.Any(), "hello world", gomock.Any()).Return(false, nil)
		txSvc.EXPECT().CreateRawTransaction(gomock.Any(), gomock.Any()).
			Return(nil, nil)

		err := sh.ExecuteTask(context.TODO(), database.ScheduleRule{
			Script: "hello world",
		})
		assert.NoError(t, err)
	})

	t.Run("rule interpreter error", func(t *testing.T) {
		ruleInt := NewMockInterpreter(gomock.NewController(t))
		txSvc := NewMockTransactionSvc(gomock.NewController(t))

		sh := rules.NewScheduler(&rules.SchedulerConfig{
			RuleInterpreter: ruleInt,
			TransactionSvc:  txSvc,
		})

		ruleInt.EXPECT().Run(gomock.Any(), "hello world", gomock.Any()).
			Return(false, assert.AnError)

		err := sh.ExecuteTask(context.TODO(), database.ScheduleRule{
			Script: "hello world",
		})
		assert.Error(t, err)
	})
}

func TestReinit(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		targetRules := []*database.ScheduleRule{
			{
				Script:         "hello world",
				Title:          "Test Rule",
				CronExpression: "*/5 * * * *",
				Enabled:        true,
			},
		}

		assert.NoError(t, gormDB.Create(&targetRules).Error)

		sh := rules.NewScheduler(&rules.SchedulerConfig{})

		assert.NoError(t, sh.Reinit(context.TODO()))
		assert.NoError(t, sh.Reinit(context.TODO())) // will close previous
	})

	t.Run("fail on second rule", func(t *testing.T) {
		assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

		targetRules := []*database.ScheduleRule{
			{
				Script:         "hello world",
				Title:          "Test Rule",
				CronExpression: "*/5 * * * *",
				Enabled:        true,
			},
			{
				Script:         "another rule",
				Title:          "Another Test Rule",
				CronExpression: "*/5 * * * *xxx",
				Enabled:        true,
			},
		}

		assert.NoError(t, gormDB.Create(&targetRules).Error)

		sh := rules.NewScheduler(&rules.SchedulerConfig{})

		err := sh.Reinit(context.TODO())
		assert.ErrorContains(t, err, "crontab parse failure")
	})

	t.Run("scheduler init err", func(t *testing.T) {
		t.Run("fail", func(t *testing.T) {
			sh := rules.NewScheduler(&rules.SchedulerConfig{
				Opts: []gocron.SchedulerOption{
					gocron.WithDistributedLocker(nil),
				},
			})

			assert.ErrorContains(t, sh.Reinit(context.TODO()), "locker must not be nil")
		})
	})
}
