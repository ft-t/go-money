package transactions

import (
	"github.com/ft-t/go-money/pkg/database"
	"time"
)

type fillResponse struct {
	Accounts map[int32]*database.Account
}

type CalculateDailyStatRequest struct {
	StartDate time.Time
	AccountID int32
}

type PossibleAccount struct {
	SourceAccounts      []*database.Account
	DestinationAccounts []*database.Account
}
