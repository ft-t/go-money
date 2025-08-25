package transactions_test

import (
	"context"
	"testing"

	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestFillWithdrawalFail(t *testing.T) {
	t.Run("invalid source amount", func(t *testing.T) {
		srv := transactions.NewService(&transactions.ServiceConfig{})

		_, err := srv.FillWithdrawal(context.TODO(), &transactionsv1.Expense{
			SourceAmount: "abcd",
		}, nil)
		assert.ErrorContains(t, err, "invalid source amount")
	})

	t.Run("invalid fx amount", func(t *testing.T) {
		srv := transactions.NewService(&transactions.ServiceConfig{})

		_, err := srv.FillWithdrawal(context.TODO(), &transactionsv1.Expense{
			SourceAmount:     "1",
			FxSourceAmount:   lo.ToPtr("abcd"),
			FxSourceCurrency: lo.ToPtr("USD"),
		}, &database.Transaction{})
		assert.ErrorContains(t, err, "invalid foreign amount")
	})

	t.Run("invalid fx source amount is positive", func(t *testing.T) {
		srv := transactions.NewService(&transactions.ServiceConfig{})

		_, err := srv.FillWithdrawal(context.TODO(), &transactionsv1.Expense{
			SourceAmount:     "1",
			FxSourceAmount:   lo.ToPtr("1"),
			FxSourceCurrency: lo.ToPtr("USD"),
		}, &database.Transaction{})
		assert.ErrorContains(t, err, "foreign amount must be negative")
	})

	t.Run("missing fx source currency", func(t *testing.T) {
		srv := transactions.NewService(&transactions.ServiceConfig{})

		_, err := srv.FillWithdrawal(context.TODO(), &transactionsv1.Expense{
			SourceAmount:   "1",
			FxSourceAmount: lo.ToPtr("-1"),
		}, &database.Transaction{})
		assert.ErrorContains(t, err, "foreign currency is required when foreign amount is provided")
	})

	t.Run("invalid dest amount", func(t *testing.T) {
		srv := transactions.NewService(&transactions.ServiceConfig{})

		_, err := srv.FillWithdrawal(context.TODO(), &transactionsv1.Expense{
			SourceAmount:      "1",
			DestinationAmount: "abcd",
		}, &database.Transaction{})
		assert.ErrorContains(t, err, "invalid destination amount")
	})

	t.Run("dest amount is positive", func(t *testing.T) {
		srv := transactions.NewService(&transactions.ServiceConfig{})

		_, err := srv.FillWithdrawal(context.TODO(), &transactionsv1.Expense{
			SourceAmount:      "1",
			DestinationAmount: "-1",
		}, &database.Transaction{})
		assert.ErrorContains(t, err, "destination amount must be positive")
	})

	t.Run("missing dest currency", func(t *testing.T) {
		srv := transactions.NewService(&transactions.ServiceConfig{})

		_, err := srv.FillWithdrawal(context.TODO(), &transactionsv1.Expense{
			SourceAmount:      "1",
			DestinationAmount: "1",
		}, &database.Transaction{})
		assert.ErrorContains(t, err, "destination currency is required when destination amount is provided")
	})
}

func TestFillDeposit(t *testing.T) {
	t.Run("invalid dest amount", func(t *testing.T) {
		srv := transactions.NewService(&transactions.ServiceConfig{})

		_, err := srv.FillDeposit(context.TODO(), &transactionsv1.Income{
			DestinationAmount: "abcd",
		}, nil)
		assert.ErrorContains(t, err, "invalid destination amount")
	})

	t.Run("invalid source amount", func(t *testing.T) {
		srv := transactions.NewService(&transactions.ServiceConfig{})

		_, err := srv.FillDeposit(context.TODO(), &transactionsv1.Income{
			DestinationAmount: "1",
			SourceAmount:      "abcd",
		}, nil)
		assert.ErrorContains(t, err, "invalid source amount")
	})
}
