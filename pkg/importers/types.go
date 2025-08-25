package importers

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/shopspring/decimal"
)

type GetSecondaryAccountRequest struct {
	InitialAmount   decimal.Decimal
	InitialCurrency string

	Accounts          map[string]*database.Account
	TargetAccountName string

	TransactionType gomoneypbv1.TransactionType
}

type GetSecondaryAccountResponse struct {
	SecondaryAccount *database.Account
	SecondaryAmount  decimal.Decimal // always positive
}
