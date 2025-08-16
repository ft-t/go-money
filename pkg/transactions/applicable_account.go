package transactions

import (
	"context"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
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
	accounts, err := s.accountSvc.GetAllAccounts(ctx)
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
		gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE,
		gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
	}
	finalRes := map[gomoneypbv1.TransactionType]*PossibleAccount{}

	assetAccounts := []gomoneypbv1.AccountType{
		gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
		gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
		gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
	}

	assetAndLiabilityAccounts := append(assetAccounts, gomoneypbv1.AccountType_ACCOUNT_TYPE_LIABILITY)

	for _, txType := range txTypes {
		res := &PossibleAccount{
			SourceAccounts:      map[int32]*database.Account{},
			DestinationAccounts: map[int32]*database.Account{},
		}
		finalRes[txType] = res

		for _, account := range accounts {
			switch txType {
			case gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS:
				if lo.Contains(assetAndLiabilityAccounts, account.Type) {
					res.SourceAccounts[account.ID] = account
					res.DestinationAccounts[account.ID] = account
				}
			case gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME:
				if account.Type == gomoneypbv1.AccountType_ACCOUNT_TYPE_INCOME {
					res.SourceAccounts[account.ID] = account
				} else if lo.Contains(assetAccounts, account.Type) {
					res.DestinationAccounts[account.ID] = account
				}
			case gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE:
				if account.Type == gomoneypbv1.AccountType_ACCOUNT_TYPE_EXPENSE {
					res.DestinationAccounts[account.ID] = account
				} else if lo.Contains(assetAndLiabilityAccounts, account.Type) {
					res.SourceAccounts[account.ID] = account
				}
			}
		}
	}

	return finalRes
}
