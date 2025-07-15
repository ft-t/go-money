package fetcher

import (
	"context"
	"encoding/json"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/currency"
	"net/http"
)

type Fetcher struct {
	apiURL     string
	httpClient *http.Client
}

func NewFetcher(
	client *http.Client,
	apiURL string,
) *Fetcher {
	return &Fetcher{
		apiURL:     apiURL,
		httpClient: client,
	}
}

func (f *Fetcher) Fetch(ctx context.Context) (*currency.RemoteRates, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.apiURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed building request")
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch exchange rates")
	}

	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var parsedResponse rateResponse
	if err = json.NewDecoder(resp.Body).Decode(&parsedResponse); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	return &currency.RemoteRates{
		Base:  parsedResponse.BaseCode,
		Rates: parsedResponse.ConversionRates,
	}, nil
}
