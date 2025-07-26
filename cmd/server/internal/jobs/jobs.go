package jobs

import (
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/go-co-op/gocron/v2"
)

type Config struct {
	Configuration          configuration.Configuration
	ExchangeRatesUpdateSvc ExchangeRatesUpdateSvc
	MaintenanceSvc         MaintenanceSvc
	Opts                   []gocron.SchedulerOption
}

type JobScheduler struct {
	scheduler gocron.Scheduler
	cfg       *Config
}

func NewJobScheduler(cfg *Config) (*JobScheduler, error) {
	scheduler, err := gocron.NewScheduler(cfg.Opts...)
	if err != nil {
		return nil, err
	}

	j := &JobScheduler{
		scheduler: scheduler,
		cfg:       cfg,
	}

	if _, err = scheduler.NewJob(
		gocron.CronJob("10 12 * * *", false), // should be in sync with sync-exchange-rates service
		gocron.NewTask(j.UpdateCurrencyRates),
	); err != nil {
		return nil, errors.Wrap(err, "failed to create currency rate updater job")
	}

	if _, err = scheduler.NewJob(
		gocron.CronJob("1 0 * * *", false), // run on day start, so it will generate daily stats for current day
		gocron.NewTask(j.FixDailyGap),
	); err != nil {
		return nil, errors.Wrap(err, "failed to create currency rate updater job")
	}

	return j, nil
}

func (j *JobScheduler) GetScheduler() gocron.Scheduler {
	return j.scheduler
}

func (j *JobScheduler) StartAsync() error {
	j.scheduler.Start()
	
	return nil
}

func (j *JobScheduler) Stop() error {
	return j.scheduler.Shutdown()
}
