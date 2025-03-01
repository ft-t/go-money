package transactions

import "github.com/ft-t/go-money/pkg/database"

type fillResponse struct {
	Accounts map[int32]*database.Account
}
