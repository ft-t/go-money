package transactions_test

import (
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHandleTransactionFail(t *testing.T) {
	stat := transactions.NewStatService()

	ctx, cancel := context.WithCancel(context.TODO())
	cancel()

	assert.ErrorContains(t, stat.HandleTransactions(ctx, gormDB.WithContext(ctx), []*database.Transaction{
		{
			TransactionDateTime: time.Now().UTC(),
			SourceAccountID:     lo.ToPtr(int32(1)),
		},
	}), "context canceled")
}

func TestBuildImpactedAccounts(t *testing.T) {
	stat := transactions.NewStatService()

	txs := []*database.Transaction{
		{
			TransactionDateTime: time.Now().UTC(),
			SourceAccountID:     lo.ToPtr(int32(1)),
		},
		{
			TransactionDateTime: time.Now().UTC().Add(-time.Hour * 24 * 2),
			SourceAccountID:     lo.ToPtr(int32(1)),
		},
	}

	impacted := stat.BuildImpactedAccounts(txs)

	assert.Len(t, impacted, 1)
	assert.EqualValues(t, txs[1].TransactionDateTime, impacted[int32(1)])
}
