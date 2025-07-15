package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	fetcher2 "github.com/ft-t/go-money/cmd/sync-exchange-rates/internal/fetcher"
	uploader2 "github.com/ft-t/go-money/cmd/sync-exchange-rates/internal/uploader"

	"net/http"
	"os"
)

func main() {
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	s3Client := s3.NewFromConfig(sdkConfig)
	uploader := uploader2.NewS3(s3Client, os.Getenv("EXCHANGE_RATES_BUCKET_NAME"))

	apiURL := os.Getenv("EXCHANGE_RATES_API_URL")

	fetcher := fetcher2.NewFetcher(http.DefaultClient, apiURL)
	baseRates, err := fetcher.Fetch(context.Background())
	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(baseRates)
	if err != nil {
		panic(err)
	}

	if err = uploader.Upload(ctx, "/latest2.json", data); err != nil {
		panic(err)
	}
}
