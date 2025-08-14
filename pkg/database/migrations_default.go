package database

import (
	"fmt"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/configuration"
)

func generateDefaultAccounts(cfg *configuration.Configuration) []string {
	var result []string

	for _, acc := range []struct {
		Name string
		Type gomoneypbv1.AccountType
	}{
		{"Default Expense", gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE},
		{"Default Income", gomoneypbv1.AccountType_ACCOUNT_TYPE_INCOME},
		{"Default Income", gomoneypbv1.AccountType_ACCOUNT_TYPE_LIABILITY},
		{"Cash", gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET},
		{"Default Adjustment", gomoneypbv1.AccountType_ACCOUNT_TYPE_ADJUSTMENT},
	} {
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
where not exists (select 1 from accounts where type = %d and flags & 1 = 1)`, acc.Name, cfg.CurrencyConfig.BaseCurrency, int(acc.Type), int(acc.Type)))
	}

	return result
}
