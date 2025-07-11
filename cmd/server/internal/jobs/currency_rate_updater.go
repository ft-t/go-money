package jobs

import (
	"context"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/currency"
	"github.com/ft-t/go-money/pkg/transactions"
	"net/http"
)

func (j *JobScheduler) currencyRateUpdater(ctx context.Context) error {
	cfg := configuration.GetConfiguration()

	sync := currency.NewSyncer(http.DefaultClient, transactions.NewBaseAmountService(), cfg.CurrencyConfig)

	return sync.Sync(ctx, j.cfg.Configuration.ExchangeRatesUrl)
}
