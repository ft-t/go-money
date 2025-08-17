package double_entry

import "github.com/ft-t/go-money/pkg/database"

type RecordRequest struct {
	Transaction   *database.Transaction
	SourceAccount *database.Account
}
