package transactions

import (
	"context"
	"database/sql"
	_ "embed"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

//go:embed scripts/update_amount_in_base_currency.sql
var updateAmountInBaseCurrency string

type BaseAmountService struct {
	baseCurrency string
}

func NewBaseAmountService(
	baseCurrency string,
) *BaseAmountService {
	return &BaseAmountService{
		baseCurrency: baseCurrency,
	}
}

func (s *BaseAmountService) RecalculateAmountInBaseCurrencyForAll(
	ctx context.Context,
	tx *gorm.DB,
) error {
	return s.RecalculateAmountInBaseCurrency(ctx, tx, nil)
}

func (s *BaseAmountService) RecalculateAmountInBaseCurrency(
	_ context.Context,
	tx *gorm.DB,
	specificTxIDs []*database.Transaction,
) error {
	var target any = specificTxIDs

	var txIDs []int64
	txMap := map[int64]*database.Transaction{}

	for _, createdTx := range specificTxIDs {
		txMap[createdTx.ID] = createdTx

		txIDs = append(txIDs, createdTx.ID)
	}

	if len(txIDs) > 0 {
		target = pq.Int64Array(txIDs)
	}

	var results []*struct {
		Id                              int64
		DestinationAmountInBaseCurrency decimal.NullDecimal
		SourceAmountInBaseCurrency      decimal.NullDecimal
	}

	if err := tx.Raw(updateAmountInBaseCurrency,
		sql.Named("baseCurrency", s.baseCurrency),
		sql.Named("specificTxIDs", target),
	).Find(&results).Error; err != nil {
		return errors.Wrap(err, "failed to recalculate")
	}

	for _, res := range results {
		createdTx, ok := txMap[res.Id]

		if !ok {
			continue // should not happen, but just in case
		}

		createdTx.DestinationAmountInBaseCurrency = res.DestinationAmountInBaseCurrency
		createdTx.SourceAmountInBaseCurrency = res.SourceAmountInBaseCurrency
	}

	return nil
}
