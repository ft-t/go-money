package history

import (
	"encoding/json"

	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/wI2L/jsondiff"
)

func Snapshot(tx *database.Transaction) (map[string]any, error) {
	raw, err := json.Marshal(toMarshallable(tx))
	if err != nil {
		return nil, errors.Wrap(err, "marshal tx")
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, errors.Wrap(err, "unmarshal tx")
	}
	for k := range excludedFields {
		delete(m, k)
	}
	return m, nil
}

type marshallableTx struct {
	ID                       int64             `json:"id"`
	SourceAmount             any               `json:"source_amount"`
	SourceCurrency           string            `json:"source_currency"`
	SourceAmountInBase       any               `json:"source_amount_in_base_currency"`
	FxSourceAmount           any               `json:"fx_source_amount"`
	FxSourceCurrency         string            `json:"fx_source_currency"`
	DestinationAmount        any               `json:"destination_amount"`
	DestinationCurrency      string            `json:"destination_currency"`
	DestinationAmountInBase  any               `json:"destination_amount_in_base_currency"`
	SourceAccountID          int32             `json:"source_account_id"`
	DestinationAccountID     int32             `json:"destination_account_id"`
	TagIDs                   []int32           `json:"tag_ids"`
	CreatedAt                any               `json:"created_at"`
	UpdatedAt                any               `json:"updated_at"`
	DeletedAt                any               `json:"deleted_at"`
	Notes                    string            `json:"notes"`
	Extra                    map[string]string `json:"extra"`
	TransactionDateTime      any               `json:"transaction_date_time"`
	TransactionDateOnly      any               `json:"transaction_date_only"`
	TransactionType          int32             `json:"transaction_type"`
	Flags                    int64             `json:"flags"`
	VoidedByTransactionID    *int64            `json:"voided_by_transaction_id"`
	Title                    string            `json:"title"`
	ReferenceNumber          *string           `json:"reference_number"`
	InternalReferenceNumbers []string          `json:"internal_reference_numbers"`
	CategoryID               *int32            `json:"category_id"`
}

func toMarshallable(tx *database.Transaction) marshallableTx {
	var srcAmt, srcAmtBase, fxAmt, dstAmt, dstAmtBase any
	if tx.SourceAmount.Valid {
		srcAmt = tx.SourceAmount.Decimal.String()
	}
	if tx.SourceAmountInBaseCurrency.Valid {
		srcAmtBase = tx.SourceAmountInBaseCurrency.Decimal.String()
	}
	if tx.FxSourceAmount.Valid {
		fxAmt = tx.FxSourceAmount.Decimal.String()
	}
	if tx.DestinationAmount.Valid {
		dstAmt = tx.DestinationAmount.Decimal.String()
	}
	if tx.DestinationAmountInBaseCurrency.Valid {
		dstAmtBase = tx.DestinationAmountInBaseCurrency.Decimal.String()
	}
	var deletedAt any
	if tx.DeletedAt.Valid {
		deletedAt = tx.DeletedAt.Time
	}
	return marshallableTx{
		ID:                       tx.ID,
		SourceAmount:             srcAmt,
		SourceCurrency:           tx.SourceCurrency,
		SourceAmountInBase:       srcAmtBase,
		FxSourceAmount:           fxAmt,
		FxSourceCurrency:         tx.FxSourceCurrency,
		DestinationAmount:        dstAmt,
		DestinationCurrency:      tx.DestinationCurrency,
		DestinationAmountInBase:  dstAmtBase,
		SourceAccountID:          tx.SourceAccountID,
		DestinationAccountID:     tx.DestinationAccountID,
		TagIDs:                   []int32(tx.TagIDs),
		CreatedAt:                tx.CreatedAt,
		UpdatedAt:                tx.UpdatedAt,
		DeletedAt:                deletedAt,
		Notes:                    tx.Notes,
		Extra:                    tx.Extra,
		TransactionDateTime:      tx.TransactionDateTime,
		TransactionDateOnly:      tx.TransactionDateOnly,
		TransactionType:          int32(tx.TransactionType),
		Flags:                    int64(tx.Flags),
		VoidedByTransactionID:    tx.VoidedByTransactionID,
		Title:                    tx.Title,
		ReferenceNumber:          tx.ReferenceNumber,
		InternalReferenceNumbers: []string(tx.InternalReferenceNumbers),
		CategoryID:               tx.CategoryID,
	}
}

func Diff(prev, curr map[string]any) (map[string]any, error) {
	patch, err := jsondiff.Compare(prev, curr)
	if err != nil {
		return nil, errors.Wrap(err, "compute json patch")
	}
	if len(patch) == 0 {
		return nil, nil
	}
	raw, err := json.Marshal(patch)
	if err != nil {
		return nil, errors.Wrap(err, "marshal patch")
	}
	var ops []any
	if err := json.Unmarshal(raw, &ops); err != nil {
		return nil, errors.Wrap(err, "unmarshal patch")
	}
	return map[string]any{"ops": ops}, nil
}
