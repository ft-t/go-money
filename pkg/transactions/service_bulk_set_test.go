package transactions_test

import (
	"context"
	"testing"
	"time"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_BulkSetCategory_Success(t *testing.T) {
	type txSeed struct {
		title      string
		categoryID *int32
	}

	type caseDef struct {
		name       string
		seeds      []txSeed
		buildAssgn func(ids []int64) []transactions.CategoryAssignment
		want       []*int32 // expected category per tx after call, in seed order
	}

	cases := []caseDef{
		{
			name: "single assignment sets category",
			seeds: []txSeed{
				{title: "tx1", categoryID: nil},
			},
			buildAssgn: func(ids []int64) []transactions.CategoryAssignment {
				return []transactions.CategoryAssignment{
					{TransactionID: ids[0], CategoryID: lo.ToPtr(int32(7))},
				}
			},
			want: []*int32{lo.ToPtr(int32(7))},
		},
		{
			name: "single assignment clears category",
			seeds: []txSeed{
				{title: "tx1", categoryID: lo.ToPtr(int32(11))},
			},
			buildAssgn: func(ids []int64) []transactions.CategoryAssignment {
				return []transactions.CategoryAssignment{
					{TransactionID: ids[0], CategoryID: nil},
				}
			},
			want: []*int32{nil},
		},
		{
			name: "multiple assignments in one call",
			seeds: []txSeed{
				{title: "tx1", categoryID: nil},
				{title: "tx2", categoryID: lo.ToPtr(int32(2))},
				{title: "tx3", categoryID: lo.ToPtr(int32(3))},
			},
			buildAssgn: func(ids []int64) []transactions.CategoryAssignment {
				return []transactions.CategoryAssignment{
					{TransactionID: ids[0], CategoryID: lo.ToPtr(int32(42))},
					{TransactionID: ids[1], CategoryID: nil},
					{TransactionID: ids[2], CategoryID: lo.ToPtr(int32(99))},
				}
			},
			want: []*int32{lo.ToPtr(int32(42)), nil, lo.ToPtr(int32(99))},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, testingutils.FlushAllTables(cfg.Db))

			txs := make([]*database.Transaction, 0, len(tc.seeds))
			for i, s := range tc.seeds {
				txs = append(txs, &database.Transaction{
					TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
					TransactionDateTime:  time.Now().Add(time.Duration(i) * time.Second),
					Title:                s.title,
					DestinationAccountID: int32(100 + i),
					DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(int64(10 + i))),
					DestinationCurrency:  "USD",
					Extra:                map[string]string{},
					CategoryID:           s.categoryID,
				})
			}

			require.NoError(t, gormDB.Create(&txs).Error)

			ids := make([]int64, 0, len(txs))
			for _, tx := range txs {
				ids = append(ids, tx.ID)
			}

			srv := transactions.NewService(&transactions.ServiceConfig{})

			err := srv.BulkSetCategory(context.TODO(), tc.buildAssgn(ids))
			require.NoError(t, err)

			var loaded []*database.Transaction
			require.NoError(t, gormDB.Where("id IN ?", ids).Order("id").Find(&loaded).Error)
			require.Len(t, loaded, len(tc.want))

			for i, got := range loaded {
				assert.Equal(t, tc.want[i], got.CategoryID, "tx index %d", i)
			}
		})
	}
}

func TestService_BulkSetCategory_Empty(t *testing.T) {
	srv := transactions.NewService(&transactions.ServiceConfig{})
	assert.NoError(t, srv.BulkSetCategory(context.TODO(), nil))
	assert.NoError(t, srv.BulkSetCategory(context.TODO(), []transactions.CategoryAssignment{}))
}

func TestService_BulkSetCategory_Failure(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))

	tx := &database.Transaction{
		TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
		TransactionDateTime:  time.Now(),
		Title:                "tx_fail",
		DestinationAccountID: 200,
		DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(5)),
		DestinationCurrency:  "USD",
		Extra:                map[string]string{},
	}
	require.NoError(t, gormDB.Create(tx).Error)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	srv := transactions.NewService(&transactions.ServiceConfig{})

	err := srv.BulkSetCategory(ctx, []transactions.CategoryAssignment{
		{TransactionID: tx.ID, CategoryID: lo.ToPtr(int32(7))},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set category on transaction")
}
