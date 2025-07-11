package jobs

import (
	"context"
	"github.com/rs/zerolog"
)

func (j *JobScheduler) UpdateCurrencyRates(ctx context.Context) error {
	ctx = zerolog.Ctx(ctx).With().Str("job", "update_exchange_rates").Logger().WithContext(ctx)
	return j.cfg.ExchangeRatesUpdateSvc.Sync(ctx, j.cfg.Configuration.ExchangeRatesUrl)
}
