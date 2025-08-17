package validation

import "github.com/ft-t/go-money/pkg/database"

type Request struct {
	Txs                    []*database.Transaction
	Accounts               map[int32]*database.Account
	SkipAccountsValidation bool
}
