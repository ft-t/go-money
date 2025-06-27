package main

import (
	"context"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/currency"
	"github.com/ft-t/go-money/pkg/transactions"
	"net/http"
	"os"
)

const (
	defaultExchangeRatesURL = "http://go-money-exchange-rates.s3-website.eu-north-1.amazonaws.com/latest.json"
)

func main() {
	fetchURL := defaultExchangeRatesURL

	if v := os.Getenv("CUSTOM_EXCHANGE_RATES_URL"); v != "" {
		fetchURL = v
	}

	cfg := configuration.GetConfiguration()

	ctx := context.TODO()

	sync := currency.NewSyncer(http.DefaultClient, transactions.NewBaseAmountService(), cfg.CurrencyConfig)
	if err := sync.Sync(ctx, fetchURL); err != nil {
		panic(err)
	}
}
