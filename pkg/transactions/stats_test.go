package transactions_test

import (
	"context"
	"testing"
	"time"

	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/stretchr/testify/assert"
)

func TestHandleTransactionFail(t *testing.T) {
	stat := transactions.NewStatService()

	ctx, cancel := context.WithCancel(context.TODO())
	cancel()

	assert.ErrorContains(t, stat.HandleTransactions(ctx, gormDB.WithContext(ctx), []*database.Transaction{
		{
			TransactionDateTime: time.Now().UTC(),
			SourceAccountID:     int32(1),
		},
	}), "context canceled")
}

func TestBuildImpactedAccounts(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		stat := transactions.NewStatService()

		txs := []*database.Transaction{
			{
				TransactionDateTime:  time.Now().UTC(),
				SourceAccountID:      int32(1),
				DestinationAccountID: 2,
			},
			{
				TransactionDateTime:  time.Now().UTC().Add(-time.Hour * 24 * 2),
				SourceAccountID:      int32(1),
				DestinationAccountID: 2,
			},
		}

		impacted := stat.BuildImpactedAccounts(txs)

		assert.Len(t, impacted, 2)
		assert.EqualValues(t, txs[1].TransactionDateTime, impacted[int32(1)])
	})
}

func TestNoData(t *testing.T) {
	stat := transactions.NewStatService()

	ctx, cancel := context.WithCancel(context.TODO())
	cancel()

	err := stat.HandleTransactions(ctx, gormDB.WithContext(ctx), nil)
	assert.NoError(t, err)
}
