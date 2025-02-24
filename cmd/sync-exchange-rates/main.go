package main

import (
	"context"
	"github.com/ft-t/go-money/pkg/currency"
	"net/http"
	"os"
)

const (
	defaultExchangeRatesURL = "https://localhost/latest.json" // todo
)

func main() {
	fetchURL := defaultExchangeRatesURL

	if v := os.Getenv("CUSTOM_EXCHANGE_RATES_URL"); v != "" {
		fetchURL = v
	}

	ctx := context.TODO()

	sync := currency.NewSyncer(http.DefaultClient)
	if err := sync.Sync(ctx, fetchURL); err != nil {
		panic(err)
	}
}
