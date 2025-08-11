package transactions

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
)

type ApplicableAccountService struct {
	accountSvc AccountSvc
}

func NewApplicableAccountService(
	svc AccountSvc,
) *ApplicableAccountService {
	return &ApplicableAccountService{
		accountSvc: svc,
	}
}

func (s *ApplicableAccountService) GetAll(
	ctx context.Context,
) (map[gomoneypbv1.TransactionType]*PossibleAccount, error) {
	accounts, err := s.accountSvc.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	return s.GetApplicableAccounts(ctx, accounts), nil
}

func (s *ApplicableAccountService) GetApplicableAccounts(
	_ context.Context,
	accounts []*database.Account,
) map[gomoneypbv1.TransactionType]*PossibleAccount {
	txTypes := []gomoneypbv1.TransactionType{
		gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS,
		gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL,
		gomoneypbv1.TransactionType_TRANSACTION_TYPE_DEPOSIT,
	}
	finalRes := map[gomoneypbv1.TransactionType]*PossibleAccount{}

	assetAccounts := []gomoneypbv1.AccountType{
		gomoneypbv1.AccountType_ACCOUNT_TYPE_REGULAR,
		gomoneypbv1.AccountType_ACCOUNT_TYPE_SAVINGS,
		gomoneypbv1.AccountType_ACCOUNT_TYPE_BROKERAGE,
	}

	assetAndLiabilityAccounts := append(assetAccounts, gomoneypbv1.AccountType_ACCOUNT_TYPE_LIABILITY)

	for _, txType := range txTypes {
		res := &PossibleAccount{}
		finalRes[txType] = res

		for _, account := range accounts {
			switch txType {
			case gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS:
				if lo.Contains(assetAndLiabilityAccounts, account.Type) {
					res.SourceAccounts = append(res.SourceAccounts, account)
					res.DestinationAccounts = append(res.DestinationAccounts, account)
				}
			case gomoneypbv1.TransactionType_TRANSACTION_TYPE_DEPOSIT:
				if account.Type == gomoneypbv1.AccountType_ACCOUNT_TYPE_INCOME {
					res.SourceAccounts = append(res.SourceAccounts, account)
				} else if lo.Contains(assetAccounts, account.Type) {
					res.DestinationAccounts = append(res.DestinationAccounts, account)
				}
			case gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL:
				if account.Type == gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE { // dest always expense
					res.DestinationAccounts = append(res.DestinationAccounts, account)
				} else if lo.Contains(assetAndLiabilityAccounts, account.Type) {
					res.SourceAccounts = append(res.SourceAccounts, account)
				}
			}
		}
	}

	return finalRes
}
