package currency

import (
	"context"
	"encoding/json"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"gorm.io/gorm/clause"
	"net/http"
	"time"
)

type Syncer struct {
	cl httpClient
}

func NewSyncer(cl httpClient) *Syncer {
	return &Syncer{cl: cl}
}

func (s *Syncer) Sync(
	ctx context.Context,
	remoteURL string,
) error {
	httpReq, httpReqErr := http.NewRequestWithContext(ctx, "GET", remoteURL, nil)
	if httpReqErr != nil {
		return httpReqErr
	}

	resp, err := s.cl.Do(httpReq)
	if err != nil {
		return err
	}

	if resp.Body != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}

	var parsed remoteRates
	if err = json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return err
	}

	tx := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster)).Begin()
	defer tx.Rollback()

	for currency, rate := range parsed.Rates {
		if err = tx.Clauses(clause.OnConflict{
			OnConstraint: "currencies_pk",
			DoUpdates: clause.Set{
				{Column: clause.Column{
					Name: "rate",
				}, Value: rate},
			},
		}).Create(&database.Currency{
			ID:            currency,
			Rate:          rate,
			DecimalPlaces: configuration.DefaultDecimalPlaces,
			UpdatedAt:     time.Now().UTC(),
		}).Error; err != nil {
			return err
		}
	}

	return tx.Commit().Error
}
