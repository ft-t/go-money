package transactions

import (
	"time"

	"github.com/ft-t/go-money/pkg/database"
)

type FillResponse struct {
	Accounts map[int32]*database.Account
}

type CalculateDailyStatRequest struct {
	StartDate time.Time
	AccountID int32
}

type UpsertOptions struct {
	SkipAccountSourceDestValidation bool
	SkipValidationErrors            bool
}

type CategoryAssignment struct {
	TransactionID int64
	CategoryID    *int32 // nil clears the category
}

type TagsAssignment struct {
	TransactionID int64
	TagIDs        []int32 // nil or empty clears all tags
}
