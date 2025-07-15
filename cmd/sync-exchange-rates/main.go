package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	fetcher2 "github.com/ft-t/go-money/cmd/sync-exchange-rates/internal/fetcher"
	uploader2 "github.com/ft-t/go-money/cmd/sync-exchange-rates/internal/uploader"
	"net/http"
	"os"
	"time"
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

	lambda.Start(func(ctx context.Context, raw json.RawMessage) error {
		return Handler(ctx, raw, fetcher, uploader)
	})
}

func Handler(
	ctx context.Context,
	_ json.RawMessage,
	fetcher *fetcher2.Fetcher,
	uploader *uploader2.S3,
) error {
	baseRates, err := fetcher.Fetch(context.Background())
	if err != nil {
		return err
	}

	baseRates.UpdatedAt = time.Now().UTC()

	data, err := json.Marshal(baseRates)
	if err != nil {
		return err
	}

	return uploader.Upload(ctx, "latest.json", data)
}
