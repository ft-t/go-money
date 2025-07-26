package rules

import "github.com/go-co-op/gocron/v2"

type Scheduler struct {
	cfg *SchedulerConfig
}

type SchedulerConfig struct {
	Opts []gocron.SchedulerOption
}

func NewScheduler(
	cfg *SchedulerConfig,
) *Scheduler {
	return &Scheduler{
		cfg: cfg,
	}
}

func (s *Scheduler) ValidateCronExpression(cron string) error {
	sh, err := gocron.NewScheduler(s.cfg.Opts...)
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
