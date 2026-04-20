package history_test

import (
	"context"
	"os"
	"testing"

	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions/history"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var (
	cfg    *configuration.Configuration
	gormDB *gorm.DB
)

func TestMain(m *testing.M) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)
	os.Exit(m.Run())
}

func TestService_Record_Created_Success(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))
	svc := history.NewService()

	curr := &database.Transaction{ID: 1, Title: "new title", SourceCurrency: "USD"}
	err := svc.Record(context.Background(), gormDB, history.RecordRequest{
		Tx:        curr,
		EventType: database.TransactionHistoryEventTypeCreated,
		Actor:     history.UserActor(7),
	})
	require.NoError(t, err)

	var rows []database.TransactionHistory
	require.NoError(t, gormDB.Where("transaction_id = ?", int64(1)).Find(&rows).Error)
	require.Len(t, rows, 1)
	assert.Equal(t, database.TransactionHistoryEventTypeCreated, rows[0].EventType)
	assert.Equal(t, database.TransactionHistoryActorTypeUser, rows[0].ActorType)
	require.NotNil(t, rows[0].ActorUserID)
	assert.Equal(t, int32(7), *rows[0].ActorUserID)
	assert.Nil(t, rows[0].Diff)
	assert.NotNil(t, rows[0].Snapshot)
	assert.Equal(t, "new title", rows[0].Snapshot["title"])
}

func TestService_Record_Updated_Success(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))
	svc := history.NewService()

	prev := &database.Transaction{ID: 2, Title: "old", SourceCurrency: "USD"}
	curr := &database.Transaction{ID: 2, Title: "new", SourceCurrency: "USD"}

	err := svc.Record(context.Background(), gormDB, history.RecordRequest{
		Tx:        curr,
		Previous:  prev,
		EventType: database.TransactionHistoryEventTypeUpdated,
		Actor:     history.UserActor(3),
	})
	require.NoError(t, err)

	var rows []database.TransactionHistory
	require.NoError(t, gormDB.Where("transaction_id = ?", int64(2)).Find(&rows).Error)
	require.Len(t, rows, 1)
	assert.Equal(t, database.TransactionHistoryEventTypeUpdated, rows[0].EventType)
	require.NotNil(t, rows[0].Diff)
	ops, ok := rows[0].Diff["ops"].([]any)
	require.True(t, ok)
	assert.Len(t, ops, 1)
}

func TestService_Record_Updated_NoChange_NilDiff(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))
	svc := history.NewService()

	tx := &database.Transaction{ID: 5, Title: "same"}
	err := svc.Record(context.Background(), gormDB, history.RecordRequest{
		Tx:        tx,
		Previous:  tx,
		EventType: database.TransactionHistoryEventTypeUpdated,
		Actor:     history.UserActor(1),
	})
	require.NoError(t, err)

	var rows []database.TransactionHistory
	require.NoError(t, gormDB.Where("transaction_id = ?", int64(5)).Find(&rows).Error)
	require.Len(t, rows, 1)
	assert.Nil(t, rows[0].Diff)
}

func TestService_Record_ActorDetail_Persisted(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))
	svc := history.NewService()

	tx := &database.Transaction{ID: 8, Title: "t"}
	err := svc.Record(context.Background(), gormDB, history.RecordRequest{
		Tx:        tx,
		EventType: database.TransactionHistoryEventTypeRuleApplied,
		Actor:     history.ImporterActor("firefly"),
	})
	require.NoError(t, err)

	var rows []database.TransactionHistory
	require.NoError(t, gormDB.Where("transaction_id = ?", int64(8)).Find(&rows).Error)
	require.Len(t, rows, 1)
	require.NotNil(t, rows[0].ActorExtra)
	assert.Equal(t, "firefly", *rows[0].ActorExtra)
}

func TestService_List_Success(t *testing.T) {
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))
	svc := history.NewService()

	tx1 := &database.Transaction{ID: 9, Title: "v1"}
	tx2 := &database.Transaction{ID: 9, Title: "v2"}

	require.NoError(t, svc.Record(context.Background(), gormDB, history.RecordRequest{
		Tx: tx1, EventType: database.TransactionHistoryEventTypeCreated, Actor: history.UserActor(1),
	}))
	require.NoError(t, svc.Record(context.Background(), gormDB, history.RecordRequest{
		Tx: tx2, Previous: tx1, EventType: database.TransactionHistoryEventTypeUpdated, Actor: history.UserActor(1),
	}))

	rows, err := svc.List(context.Background(), 9)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, database.TransactionHistoryEventTypeCreated, rows[0].EventType)
	assert.Equal(t, database.TransactionHistoryEventTypeUpdated, rows[1].EventType)
}
