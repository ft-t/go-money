package rules

import (
	"context"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/go-co-op/gocron/v2"
)

type Scheduler struct {
	cfg       *SchedulerConfig
	scheduler gocron.Scheduler
}

type SchedulerConfig struct {
	Opts               []gocron.SchedulerOption
	CronValidationOpts []gocron.SchedulerOption
	RuleInterpreter    Interpreter
	TransactionSvc     TransactionSvc
}

func NewScheduler(
	cfg *SchedulerConfig,
) *Scheduler {
	return &Scheduler{
		cfg: cfg,
	}
}

func (s *Scheduler) Reinit(ctx context.Context) error {
	var rules []database.ScheduleRule

	if err := database.GetDbWithContext(ctx, database.DbTypeReadonly).
		Where("enabled = true").
		Find(&rules).Error; err != nil {
		return err
	}

	sh, err := gocron.NewScheduler(s.cfg.Opts...)
	if err != nil {
		return err
	}

	for _, rule := range rules {
		ruleRef := rule

		_, err = sh.NewJob(
			gocron.CronJob(rule.CronExpression, false),
			gocron.NewTask(func(
				innerCtx context.Context,
			) error {
				return s.ExecuteTask(innerCtx, ruleRef)
			}),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create job for rule_id: %d", rule.ID)
		}
	}

	if s.scheduler != nil {
		_ = s.scheduler.Shutdown()
	}

	s.scheduler = sh
	sh.Start()

	return nil
}

func (s *Scheduler) ExecuteTask(
	ctx context.Context,
	rule database.ScheduleRule,
) error {
	tx := &database.Transaction{}

	_, err := s.cfg.RuleInterpreter.Run(ctx, rule.Script, tx)
	if err != nil {
		return errors.Wrapf(err, "failed to run rule script for rule_id: %d", rule.ID)
	}

	_, err = s.cfg.TransactionSvc.CreateRawTransaction(ctx, tx)

	return err
}

func (s *Scheduler) ValidateCronExpression(cron string) error {
	sh, err := gocron.NewScheduler(s.cfg.CronValidationOpts...)
	if err != nil {
		return err
	}

	if _, err = sh.NewJob(
		gocron.CronJob(cron, false),
		gocron.NewTask(s.SchedulerTestTask),
	); err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) SchedulerTestTask() error {
	return nil
}
