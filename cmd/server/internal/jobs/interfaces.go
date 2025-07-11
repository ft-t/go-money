package jobs

import "context"

//go:generate mockgen -destination interfaces_mocks_test.go -package jobs_test -source=interfaces.go

type MaintenanceSvc interface {
	FixDailyGaps(
		ctx context.Context,
	) error
}

type ExchangeRatesUpdateSvc interface {
	Sync(
		ctx context.Context,
		remoteURL string,
	) error
}
