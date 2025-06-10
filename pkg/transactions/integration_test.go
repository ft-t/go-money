package transactions_test

import (
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestBasicCalc(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	statsSvc := transactions.NewStatService()

	srv := transactions.NewService(&transactions.ServiceConfig{
		StatsSvc: statsSvc,
	})

	accounts := []*database.Account{
		{
			Name:     "BNP Paribas [USD]",
			Currency: "USD",
			Extra:    map[string]string{},
		},
		{
			Name:     "BNP Paribas [PLN]",
			Currency: "PLN",
			Extra:    map[string]string{},
		},
	}
	assert.NoError(t, gormDB.Create(&accounts).Error)

	txDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	_, err := srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(txDate),
		Transaction: &transactionsv1.CreateTransactionRequest_Deposit{
			Deposit: &transactionsv1.Deposit{
				DestinationAmount:    "500",
				DestinationCurrency:  "USD",
				DestinationAccountId: accounts[0].ID,
			},
		},
	})
	assert.NoError(t, err)

	txTransferDate := time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC)
	_, err = srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(txTransferDate),
		Transaction: &transactionsv1.CreateTransactionRequest_TransferBetweenAccounts{
			TransferBetweenAccounts: &transactionsv1.TransferBetweenAccounts{
				SourceAccountId:      accounts[0].ID,
				SourceAmount:         "-120",
				SourceCurrency:       "USD",
				DestinationAccountId: accounts[1].ID,
				DestinationAmount:    "480",
				DestinationCurrency:  "PLN",
			},
		},
	})

	expenseDate := time.Date(2025, 6, 3, 0, 0, 0, 0, time.UTC)
	_, err = srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(expenseDate),
		Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
			Withdrawal: &transactionsv1.Withdrawal{
				SourceAmount:    "-10",
				SourceCurrency:  "USD",
				SourceAccountId: accounts[0].ID,
			},
		},
	})

	assert.NoError(t, err)

	expenseDate2 := time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC)
	_, err = srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(expenseDate2),
		Transaction: &transactionsv1.CreateTransactionRequest_Withdrawal{
			Withdrawal: &transactionsv1.Withdrawal{
				SourceAmount:    "-15",
				SourceCurrency:  "PLN",
				SourceAccountId: accounts[1].ID,
			},
		},
	})

	depositDate2 := time.Date(2025, 6, 7, 0, 0, 0, 0, time.UTC)
	_, err = srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(depositDate2),
		Transaction: &transactionsv1.CreateTransactionRequest_Deposit{
			Deposit: &transactionsv1.Deposit{
				DestinationAmount:    "11",
				DestinationCurrency:  "PLN",
				DestinationAccountId: accounts[1].ID,
			},
		},
	})

	depositDate3 := time.Date(2025, 6, 9, 0, 0, 0, 0, time.UTC)
	_, err = srv.Create(context.TODO(), &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(depositDate3),
		Transaction: &transactionsv1.CreateTransactionRequest_Deposit{
			Deposit: &transactionsv1.Deposit{
				DestinationAmount:    "55",
				DestinationCurrency:  "PLN",
				DestinationAccountId: accounts[1].ID,
			},
		},
	})

	assert.NoError(t, err)
}
