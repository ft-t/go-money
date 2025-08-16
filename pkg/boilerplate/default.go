package boilerplate

import gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"

var RequiredDefaultAccounts = map[string]gomoneypbv1.AccountType{
	"Default Expense":    gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE,
	"Default Income":     gomoneypbv1.AccountType_ACCOUNT_TYPE_INCOME,
	"Default Liability":  gomoneypbv1.AccountType_ACCOUNT_TYPE_LIABILITY,
	"Cash":               gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
	"Default Adjustment": gomoneypbv1.AccountType_ACCOUNT_TYPE_ADJUSTMENT,
}
