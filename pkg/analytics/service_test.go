package analytics_test

import (
	"context"
	"os"
	"testing"
	"time"

	analyticsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/analytics/v1"
	"github.com/ft-t/go-money/pkg/analytics"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

var gormDB *gorm.DB
var cfg *configuration.Configuration

func TestMain(m *testing.M) {
	cfg = configuration.GetConfiguration()
	gormDB = database.GetDb(database.DbTypeMaster)

	os.Exit(m.Run())
}

func TestService_GetDebitsAndCreditsSummary(t *testing.T) {
	assert.NoError(t, testingutils.FlushAllTables(cfg.Db))

	account1ID := int32(1)
	account2ID := int32(2)

	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	doubleEntries := []*database.DoubleEntry{
		{
			TransactionID:        1,
			IsDebit:              false,
			AmountInBaseCurrency: decimal.NewFromInt(1000),
			BaseCurrency:         "USD",
			AccountID:            account1ID,
			CreatedAt:            baseTime.Add(-2 * time.Hour),
		},
		{
			TransactionID:        2,
			IsDebit:              true,
			AmountInBaseCurrency: decimal.NewFromInt(500),
			BaseCurrency:         "USD",
			AccountID:            account1ID,
			CreatedAt:            baseTime.Add(-1 * time.Hour),
		},
		{
			TransactionID:        3,
			IsDebit:              true,
			AmountInBaseCurrency: decimal.NewFromInt(200),
			BaseCurrency:         "USD",
			AccountID:            account1ID,
			CreatedAt:            baseTime,
		},
		{
			TransactionID:        3,
			IsDebit:              false,
			AmountInBaseCurrency: decimal.NewFromInt(200),
			BaseCurrency:         "USD",
			AccountID:            account2ID,
			CreatedAt:            baseTime,
		},
		{
			TransactionID:        4,
			IsDebit:              false,
			AmountInBaseCurrency: decimal.NewFromInt(300),
			BaseCurrency:         "USD",
			AccountID:            account2ID,
			CreatedAt:            baseTime.Add(1 * time.Hour),
		},
	}

	for _, entry := range doubleEntries {
		assert.NoError(t, gormDB.Create(entry).Error)
	}

	service := analytics.NewService()

	t.Run("calculate summary for account 1", func(t *testing.T) {
		req := &analyticsv1.GetDebitsAndCreditsSummaryRequest{
			AccountIds: []int32{account1ID},
			StartAt:    timestamppb.New(baseTime.Add(-3 * time.Hour)),
			EndAt:      timestamppb.New(baseTime.Add(3 * time.Hour)),
		}

		resp, err := service.GetDebitsAndCreditsSummary(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)


		assert.NotNil(t, resp.Items)
		assert.Contains(t, resp.Items, account1ID)

		item := resp.Items[account1ID]
		assert.Equal(t, int32(700), item.TotalDebitsAmount)
		assert.Equal(t, int32(1000), item.TotalCreditsAmount)
		assert.Equal(t, int32(2), item.TotalDebitsCount)
		assert.Equal(t, int32(1), item.TotalCreditsCount)
	})

	t.Run("calculate summary for account 2", func(t *testing.T) {
		req := &analyticsv1.GetDebitsAndCreditsSummaryRequest{
			AccountIds: []int32{account2ID},
			StartAt:    timestamppb.New(baseTime.Add(-3 * time.Hour)),
			EndAt:      timestamppb.New(baseTime.Add(3 * time.Hour)),
		}

		resp, err := service.GetDebitsAndCreditsSummary(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)


		assert.NotNil(t, resp.Items)
		assert.Contains(t, resp.Items, account2ID)

		item := resp.Items[account2ID]
		assert.Equal(t, int32(0), item.TotalDebitsAmount)
		assert.Equal(t, int32(500), item.TotalCreditsAmount)
		assert.Equal(t, int32(0), item.TotalDebitsCount)
		assert.Equal(t, int32(2), item.TotalCreditsCount)
	})

	t.Run("calculate summary with date range filter", func(t *testing.T) {
		req := &analyticsv1.GetDebitsAndCreditsSummaryRequest{
			AccountIds: []int32{account1ID},
			StartAt:    timestamppb.New(baseTime.Add(-30 * time.Minute)),
			EndAt:      timestamppb.New(baseTime.Add(30 * time.Minute)),
		}

		resp, err := service.GetDebitsAndCreditsSummary(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)


		assert.NotNil(t, resp.Items)
		assert.Contains(t, resp.Items, account1ID)

		item := resp.Items[account1ID]
		assert.Equal(t, int32(200), item.TotalDebitsAmount)
		assert.Equal(t, int32(0), item.TotalCreditsAmount)
		assert.Equal(t, int32(1), item.TotalDebitsCount)
		assert.Equal(t, int32(0), item.TotalCreditsCount)
	})

	t.Run("empty account ids", func(t *testing.T) {
		req := &analyticsv1.GetDebitsAndCreditsSummaryRequest{
			AccountIds: []int32{},
		}

		resp, err := service.GetDebitsAndCreditsSummary(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "account_ids cannot be empty")
	})

	t.Run("missing start_at", func(t *testing.T) {
		req := &analyticsv1.GetDebitsAndCreditsSummaryRequest{
			AccountIds: []int32{account1ID},
			EndAt:      timestamppb.New(baseTime.Add(1 * time.Hour)),
		}

		resp, err := service.GetDebitsAndCreditsSummary(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "start_at is required")
	})

	t.Run("missing end_at", func(t *testing.T) {
		req := &analyticsv1.GetDebitsAndCreditsSummaryRequest{
			AccountIds: []int32{account1ID},
			StartAt:    timestamppb.New(baseTime.Add(-1 * time.Hour)),
		}

		resp, err := service.GetDebitsAndCreditsSummary(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "end_at is required")
	})

	t.Run("invalid date range", func(t *testing.T) {
		req := &analyticsv1.GetDebitsAndCreditsSummaryRequest{
			AccountIds: []int32{account1ID},
			StartAt:    timestamppb.New(baseTime.Add(1 * time.Hour)),
			EndAt:      timestamppb.New(baseTime.Add(-1 * time.Hour)),
		}

		resp, err := service.GetDebitsAndCreditsSummary(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "start_at cannot be after end_at")
	})

	t.Run("no transactions found", func(t *testing.T) {
		req := &analyticsv1.GetDebitsAndCreditsSummaryRequest{
			AccountIds: []int32{int32(999)},
			StartAt:    timestamppb.New(baseTime.Add(-3 * time.Hour)),
			EndAt:      timestamppb.New(baseTime.Add(3 * time.Hour)),
		}

		resp, err := service.GetDebitsAndCreditsSummary(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		assert.NotNil(t, resp.Items)
		assert.Contains(t, resp.Items, int32(999))

		item := resp.Items[int32(999)]
		assert.Equal(t, int32(0), item.TotalDebitsAmount)
		assert.Equal(t, int32(0), item.TotalCreditsAmount)
		assert.Equal(t, int32(0), item.TotalDebitsCount)
		assert.Equal(t, int32(0), item.TotalCreditsCount)
	})
}
