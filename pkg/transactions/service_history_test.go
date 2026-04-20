package transactions_test

import (
	"context"
	"testing"
	"time"

	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/ft-t/go-money/pkg/transactions/history"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

func newHistoryTestSvc(
	t *testing.T,
	historySvc transactions.HistorySvc,
	accounts []*database.Account,
	mapperTimes int,
) *transactions.Service {
	t.Helper()

	statsSvc := transactions.NewStatService()
	mapper := NewMockMapperSvc(gomock.NewController(t))
	baseCurrency := NewMockBaseAmountSvc(gomock.NewController(t))
	ruleEngine := NewMockRuleSvc(gomock.NewController(t))
	accountSvc := NewMockAccountSvc(gomock.NewController(t))
	validationSvc := NewMockValidationSvc(gomock.NewController(t))
	doubleEntry := NewMockDoubleEntrySvc(gomock.NewController(t))

	baseCurrency.EXPECT().RecalculateAmountInBaseCurrency(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	ruleEngine.EXPECT().ProcessTransactions(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, i []*database.Transaction) ([]*database.Transaction, error) {
			return i, nil
		}).AnyTimes()
	accountSvc.EXPECT().GetAllAccounts(gomock.Any()).Return(accounts, nil).AnyTimes()
	validationSvc.EXPECT().Validate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	doubleEntry.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mapper.EXPECT().MapTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, tx *database.Transaction) *gomoneypbv1.Transaction {
			return &gomoneypbv1.Transaction{Id: tx.ID}
		}).Times(mapperTimes)

	return transactions.NewService(&transactions.ServiceConfig{
		StatsSvc:          statsSvc,
		MapperSvc:         mapper,
		BaseAmountService: baseCurrency,
		RuleSvc:           ruleEngine,
		AccountSvc:        accountSvc,
		ValidationSvc:     validationSvc,
		DoubleEntry:       doubleEntry,
		HistorySvc:        historySvc,
	})
}

func incomeRequest(account *database.Account, when time.Time) *transactionsv1.CreateTransactionRequest {
	return &transactionsv1.CreateTransactionRequest{
		TransactionDate: timestamppb.New(when),
		Transaction: &transactionsv1.CreateTransactionRequest_Income{
			Income: &transactionsv1.Income{
				DestinationAmount:    "100",
				DestinationCurrency:  account.Currency,
				DestinationAccountId: account.ID,
				SourceAmount:         "-100",
				SourceCurrency:       account.Currency,
				SourceAccountId:      account.ID,
			},
		},
	}
}

func seedAccount(t *testing.T) *database.Account {
	t.Helper()
	require.NoError(t, testingutils.FlushAllTables(cfg.Db))
	acc := &database.Account{
		Name:     "History Acc [USD]",
		Currency: "USD",
		Extra:    map[string]string{},
		Type:     gomoneypbv1.AccountType_ACCOUNT_TYPE_ASSET,
	}
	require.NoError(t, gormDB.Create(acc).Error)
	return acc
}

func TestCreate_RecordsHistoryWithUserActor(t *testing.T) {
	acc := seedAccount(t)

	historyMock := NewMockHistorySvc(gomock.NewController(t))
	srv := newHistoryTestSvc(t, historyMock, []*database.Account{acc}, 1)

	historyMock.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ *gorm.DB, req history.RecordRequest) error {
			assert.Equal(t, database.TransactionHistoryEventTypeCreated, req.EventType)
			assert.Nil(t, req.Previous)
			require.NotNil(t, req.Tx)
			assert.NotZero(t, req.Tx.ID)
			assert.Equal(t, database.TransactionHistoryActorTypeUser, req.Actor.Type)
			require.NotNil(t, req.Actor.UserID)
			assert.Equal(t, int32(7), *req.Actor.UserID)
			return nil
		}).Times(1)

	ctx := history.WithActor(context.Background(), history.UserActor(7))
	resp, err := srv.Create(ctx, incomeRequest(acc, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)))
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestUpdate_RecordsHistoryWithPrevious(t *testing.T) {
	acc := seedAccount(t)

	existing := &database.Transaction{
		TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
		TransactionDateTime:  time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		Title:                "old title",
		DestinationAccountID: acc.ID,
		DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(50)),
		DestinationCurrency:  acc.Currency,
		SourceAccountID:      acc.ID,
		SourceAmount:         decimal.NewNullDecimal(decimal.NewFromInt(-50)),
		SourceCurrency:       acc.Currency,
		Extra:                map[string]string{},
	}
	require.NoError(t, gormDB.Create(existing).Error)
	existingID := existing.ID

	historyMock := NewMockHistorySvc(gomock.NewController(t))
	srv := newHistoryTestSvc(t, historyMock, []*database.Account{acc}, 1)

	historyMock.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ *gorm.DB, req history.RecordRequest) error {
			assert.Equal(t, database.TransactionHistoryEventTypeUpdated, req.EventType)
			require.NotNil(t, req.Previous)
			assert.Equal(t, existingID, req.Previous.ID)
			assert.Equal(t, "old title", req.Previous.Title)
			require.NotNil(t, req.Tx)
			assert.Equal(t, existingID, req.Tx.ID)
			assert.Equal(t, database.TransactionHistoryActorTypeUser, req.Actor.Type)
			return nil
		}).Times(1)

	ctx := history.WithActor(context.Background(), history.UserActor(11))
	_, err := srv.Update(ctx, &transactionsv1.UpdateTransactionRequest{
		Id:          existingID,
		Transaction: incomeRequest(acc, time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC)),
	})
	require.NoError(t, err)
}

func TestDelete_RecordsHistory(t *testing.T) {
	acc := seedAccount(t)

	tx := &database.Transaction{
		TransactionType:      gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME,
		TransactionDateTime:  time.Now(),
		Title:                "to delete",
		DestinationAccountID: acc.ID,
		DestinationAmount:    decimal.NewNullDecimal(decimal.NewFromInt(20)),
		DestinationCurrency:  acc.Currency,
		Extra:                map[string]string{},
	}
	require.NoError(t, gormDB.Create(tx).Error)
	txID := tx.ID

	ctrl := gomock.NewController(t)
	doubleEntry := NewMockDoubleEntrySvc(ctrl)
	statsSvc := NewMockStatsSvc(ctrl)
	historyMock := NewMockHistorySvc(ctrl)

	doubleEntry.EXPECT().DeleteByTransactionIDs(gomock.Any(), gomock.Any(), gomock.Eq([]int64{txID})).Return(nil)
	statsSvc.EXPECT().HandleTransactions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	historyMock.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ *gorm.DB, req history.RecordRequest) error {
			assert.Equal(t, database.TransactionHistoryEventTypeDeleted, req.EventType)
			assert.Nil(t, req.Previous)
			require.NotNil(t, req.Tx)
			assert.Equal(t, txID, req.Tx.ID)
			assert.Equal(t, database.TransactionHistoryActorTypeUser, req.Actor.Type)
			require.NotNil(t, req.Actor.UserID)
			assert.Equal(t, int32(42), *req.Actor.UserID)
			return nil
		}).Times(1)

	srv := transactions.NewService(&transactions.ServiceConfig{
		DoubleEntry: doubleEntry,
		StatsSvc:    statsSvc,
		HistorySvc:  historyMock,
	})

	ctx := history.WithActor(context.Background(), history.UserActor(42))
	resp, err := srv.DeleteTransaction(ctx, &transactionsv1.DeleteTransactionsRequest{Ids: []int64{txID}})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.EqualValues(t, 1, resp.DeletedCount)
}

func TestCreate_NoActor_SkipsHistory_StillSucceeds(t *testing.T) {
	acc := seedAccount(t)

	historyMock := NewMockHistorySvc(gomock.NewController(t))
	srv := newHistoryTestSvc(t, historyMock, []*database.Account{acc}, 1)

	historyMock.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	resp, err := srv.Create(context.Background(), incomeRequest(acc, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)))
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestCreate_HistoryError_DoesNotFailCreate(t *testing.T) {
	acc := seedAccount(t)

	historyMock := NewMockHistorySvc(gomock.NewController(t))
	srv := newHistoryTestSvc(t, historyMock, []*database.Account{acc}, 1)

	historyMock.EXPECT().Record(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("history backend down")).Times(1)

	ctx := history.WithActor(context.Background(), history.UserActor(7))
	resp, err := srv.Create(ctx, incomeRequest(acc, time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)))
	require.NoError(t, err)
	require.NotNil(t, resp)
}
