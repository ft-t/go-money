package jobs

import (
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/go-co-op/gocron/v2"
)

type Config struct {
	Configuration configuration.Configuration
}

type JobScheduler struct {
	scheduler gocron.Scheduler
	cfg       *Config
}

func NewJobScheduler(cfg *Config) (*JobScheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	j := &JobScheduler{
		scheduler: scheduler,
		cfg:       cfg,
	}

	if _, err = scheduler.NewJob(
		gocron.CronJob("20 0 * * *", false), // Rates are updated daily at 00:10 UTC by main server
		gocron.NewTask(j.currencyRateUpdater),
	); err != nil {
		return nil, errors.Wrap(err, "failed to create currency rate updater job")
	}

	return j, nil
}

func (j *JobScheduler) StartAsync() error {
	go func() {
		j.scheduler.Start()
	}()

	return nil
}

func (j *JobScheduler) Stop() error {
	return j.scheduler.Shutdown()
}
