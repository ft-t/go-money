package applicable_accounts

import "github.com/ft-t/go-money/pkg/database"

type PossibleAccount struct {
	SourceAccounts      map[int32]*database.Account
	DestinationAccounts map[int32]*database.Account
}
