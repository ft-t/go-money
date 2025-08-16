package database

import (
	"fmt"

	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/configuration"
)

func generateDefaultAccounts(cfg *configuration.Configuration) []string {
	var result []string

	for name, acc := range boilerplate.RequiredDefaultAccounts {
		result = append(result, fmt.Sprintf(`insert into accounts(name, current_balance, currency, flags, last_updated_at, created_at, deleted_at, type, note,
                     account_number, iban, liability_percent, display_order, first_transaction_at)
select '%s',
       0,
       '%s',
       1,
       now(),
       now(),
       null,
       %d,
       '',
       '',
       '',
       0,
       1,
       null
where not exists (select 1 from accounts where type = %d and flags & 1 = 1)`, name, cfg.CurrencyConfig.BaseCurrency, int(acc), int(acc)))
	}

	return result
}
