package history

import "github.com/ft-t/go-money/pkg/database"

var excludedFields = map[string]struct{}{
	"id":                                  {},
	"created_at":                          {},
	"updated_at":                          {},
	"deleted_at":                          {},
	"source_amount_in_base_currency":      {},
	"destination_amount_in_base_currency": {},
}

type RecordRequest struct {
	Tx        *database.Transaction
	Previous  *database.Transaction
	EventType database.TransactionHistoryEventType
	Actor     Actor
}
