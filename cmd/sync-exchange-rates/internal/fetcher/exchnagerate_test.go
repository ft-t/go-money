package fetcher_test

import (
	"context"
	_ "embed"
	fetcher2 "github.com/ft-t/go-money/cmd/sync-exchange-rates/internal/fetcher"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

//go:embed testdata/USD.json
var rates []byte

func TestFetch(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		targetURL := "https://some-api.com/rates.json"

		httpmock.Activate(t)
		defer httpmock.Deactivate()

		fetcher := fetcher2.NewFetcher(http.DefaultClient, targetURL)

		httpmock.RegisterResponder("GET", targetURL,
			httpmock.NewStringResponder(200, string(rates)))

		resp, err := fetcher.Fetch(context.TODO())
		assert.NoError(t, err)

		assert.EqualValues(t, "USD", resp.Base)
		assert.Len(t, resp.Rates, 163)
	})

	t.Run("fail", func(t *testing.T) {
		targetURL := "https://some-api.com/rates.json"

		httpmock.Activate(t)
		defer httpmock.Deactivate()

		fetcher := fetcher2.NewFetcher(http.DefaultClient, targetURL)

		httpmock.RegisterResponder("GET", targetURL,
			httpmock.NewErrorResponder(assert.AnError))

		resp, err := fetcher.Fetch(context.TODO())
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("invalid response", func(t *testing.T) {
		targetURL := "https://some-api.com/rates.json"

		httpmock.Activate(t)
		defer httpmock.Deactivate()

		fetcher := fetcher2.NewFetcher(http.DefaultClient, targetURL)

		httpmock.RegisterResponder("GET", targetURL,
			httpmock.NewStringResponder(200, "invalid json"))

		resp, err := fetcher.Fetch(context.TODO())
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("invalid  status code", func(t *testing.T) {
		targetURL := "https://some-api.com/rates.json"

		httpmock.Activate(t)
		defer httpmock.Deactivate()

		fetcher := fetcher2.NewFetcher(http.DefaultClient, targetURL)

		httpmock.RegisterResponder("GET", targetURL,
			httpmock.NewStringResponder(500, "Internal Server Error"))

		resp, err := fetcher.Fetch(context.TODO())
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
