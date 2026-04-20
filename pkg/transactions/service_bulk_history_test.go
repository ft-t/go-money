package transactions_test

import (
	"context"
	"testing"
	"time"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/ft-t/go-money/pkg/transactions/history"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func seedBulkTxs(t *testing.T, count int) []*database.Transaction {
	t.Helper()
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))

	txs := make([]*database.Transaction, 0, count)
	for i := 0; i < count; i++ {
		txs = append(txs, &database.Transaction{
			TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
			TransactionDateTime:  time.Now().Add(time.Duration(i) * time.Second),
			Title:                "bulk tx",
			DestinationAccountID: int32(500 + i),
			DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(int64(10 + i))),
			DestinationCurrency:  "USD",
			Extra:                map[string]string{},
		})
	}
	require.NoError(t, gormDB.Create(&txs).Error)
	return txs
}

func TestBulkSetCategory_RecordsHistory_WithBulkActor(t *testing.T) {
	txs := seedBulkTxs(t, 2)

	historyMock := NewMockHistorySvc(gomock.NewController(t))
	srv := transactions.NewService(&transactions.ServiceConfig{
		HistorySvc: historyMock,
	})

	var recorded []history.RecordRequest
	historyMock.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ *gorm.DB, req history.RecordRequest) error {
			recorded = append(recorded, req)
			return nil
		}).Times(2)

	ctx := history.WithActor(context.Background(), history.UserActor(7))
	err := srv.BulkSetCategory(ctx, []transactions.CategoryAssignment{
		{TransactionID: txs[0].ID, CategoryID: lo.ToPtr(int32(42))},
		{TransactionID: txs[1].ID, CategoryID: lo.ToPtr(int32(99))},
	})
	require.NoError(t, err)

	require.Len(t, recorded, 2)
	for i, req := range recorded {
		assert.Equal(t, database.TransactionHistoryEventTypeUpdated, req.EventType)
		assert.Equal(t, database.TransactionHistoryActorTypeBulk, req.Actor.Type)
		require.NotNil(t, req.Actor.UserID, "idx %d", i)
		assert.Equal(t, int32(7), *req.Actor.UserID)
		assert.Equal(t, "set_category", req.Actor.Detail)
		require.NotNil(t, req.Previous, "idx %d", i)
		assert.Equal(t, txs[i].ID, req.Previous.ID)
		require.NotNil(t, req.Tx, "idx %d", i)
		assert.Equal(t, txs[i].ID, req.Tx.ID)
	}

	require.NotNil(t, recorded[0].Tx.CategoryID)
	assert.Equal(t, int32(42), *recorded[0].Tx.CategoryID)
	require.NotNil(t, recorded[1].Tx.CategoryID)
	assert.Equal(t, int32(99), *recorded[1].Tx.CategoryID)
}

func TestBulkSetTags_RecordsHistory_WithBulkActor(t *testing.T) {
	txs := seedBulkTxs(t, 2)

	historyMock := NewMockHistorySvc(gomock.NewController(t))
	srv := transactions.NewService(&transactions.ServiceConfig{
		HistorySvc: historyMock,
	})

	var recorded []history.RecordRequest
	historyMock.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ *gorm.DB, req history.RecordRequest) error {
			recorded = append(recorded, req)
			return nil
		}).Times(2)

	ctx := history.WithActor(context.Background(), history.UserActor(11))
	err := srv.BulkSetTags(ctx, []transactions.TagsAssignment{
		{TransactionID: txs[0].ID, TagIDs: []int32{1, 2}},
		{TransactionID: txs[1].ID, TagIDs: []int32{9}},
	})
	require.NoError(t, err)

	require.Len(t, recorded, 2)
	for i, req := range recorded {
		assert.Equal(t, database.TransactionHistoryEventTypeUpdated, req.EventType)
		assert.Equal(t, database.TransactionHistoryActorTypeBulk, req.Actor.Type)
		require.NotNil(t, req.Actor.UserID, "idx %d", i)
		assert.Equal(t, int32(11), *req.Actor.UserID)
		assert.Equal(t, "set_tags", req.Actor.Detail)
		require.NotNil(t, req.Previous, "idx %d", i)
		assert.Equal(t, txs[i].ID, req.Previous.ID)
		require.NotNil(t, req.Tx, "idx %d", i)
		assert.Equal(t, txs[i].ID, req.Tx.ID)
	}

	assert.Equal(t, pq.Int32Array{1, 2}, recorded[0].Tx.TagIDs)
	assert.Equal(t, pq.Int32Array{9}, recorded[1].Tx.TagIDs)
}

func TestBulkSetCategory_NoActor_SkipsHistory_StillSucceeds(t *testing.T) {
	txs := seedBulkTxs(t, 1)

	historyMock := NewMockHistorySvc(gomock.NewController(t))
	srv := transactions.NewService(&transactions.ServiceConfig{
		HistorySvc: historyMock,
	})

	historyMock.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err := srv.BulkSetCategory(context.Background(), []transactions.CategoryAssignment{
		{TransactionID: txs[0].ID, CategoryID: lo.ToPtr(int32(5))},
	})
	require.NoError(t, err)

	var loaded database.Transaction
	require.NoError(t, gormDB.Where("id = ?", txs[0].ID).First(&loaded).Error)
	require.NotNil(t, loaded.CategoryID)
	assert.Equal(t, int32(5), *loaded.CategoryID)
}

func TestBulkSetCategory_HistoryError_DoesNotFail(t *testing.T) {
	txs := seedBulkTxs(t, 1)

	historyMock := NewMockHistorySvc(gomock.NewController(t))
	srv := transactions.NewService(&transactions.ServiceConfig{
		HistorySvc: historyMock,
	})

	historyMock.EXPECT().
		Record(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("record failed")).
		Times(1)

	ctx := history.WithActor(context.Background(), history.UserActor(7))
	err := srv.BulkSetCategory(ctx, []transactions.CategoryAssignment{
		{TransactionID: txs[0].ID, CategoryID: lo.ToPtr(int32(42))},
	})
	require.NoError(t, err)

	var loaded database.Transaction
	require.NoError(t, gormDB.Where("id = ?", txs[0].ID).First(&loaded).Error)
	require.NotNil(t, loaded.CategoryID)
	assert.Equal(t, int32(42), *loaded.CategoryID)
}

func TestBulkSetCategory_Empty_NoHistory(t *testing.T) {
	historyMock := NewMockHistorySvc(gomock.NewController(t))
	srv := transactions.NewService(&transactions.ServiceConfig{
		HistorySvc: historyMock,
	})

	historyMock.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	ctx := history.WithActor(context.Background(), history.UserActor(1))
	require.NoError(t, srv.BulkSetCategory(ctx, nil))
	require.NoError(t, srv.BulkSetCategory(ctx, []transactions.CategoryAssignment{}))
}
