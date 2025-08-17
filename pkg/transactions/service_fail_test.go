package transactions_test

import (
	"context"
	"testing"

	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	"github.com/ft-t/go-money/pkg/transactions"
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
}
