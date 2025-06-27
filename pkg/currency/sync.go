package currency

import (
	"context"
	_ "embed"
	"encoding/json"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"gorm.io/gorm/clause"
	"net/http"
	"time"
)

type Syncer struct {
	cl            httpClient
	cfg           configuration.CurrencyConfig
	baseAmountSvc BaseAmountSvc
}

func NewSyncer(
	cl httpClient,
	baseAmountSvc BaseAmountSvc,
	config configuration.CurrencyConfig,
) *Syncer {
	return &Syncer{
		cl:            cl,
		baseAmountSvc: baseAmountSvc,
		cfg:           config,
	}
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
				{
					Column: clause.Column{
						Name: "rate",
					},
					Value: rate,
				},
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

	if s.cfg.UpdateTransactionAmountInBaseCurrency {
		if err = s.baseAmountSvc.RecalculateAmountInBaseCurrencyForAll(ctx, tx); err != nil {
			return errors.Wrap(err, "failed to recalculate")
		}
	}

	return tx.Commit().Error
}
