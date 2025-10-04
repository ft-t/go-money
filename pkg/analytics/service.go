package analytics

import (
	"context"
	"time"

	analyticsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/analytics/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/shopspring/decimal"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetDebitsAndCreditsSummary(
	ctx context.Context,
	req *analyticsv1.GetDebitsAndCreditsSummaryRequest,
) (*analyticsv1.GetDebitsAndCreditsSummaryResponse, error) {
	if len(req.AccountIds) == 0 {
		return nil, errors.New("account_ids cannot be empty")
	}

	if req.StartAt == nil {
		return nil, errors.New("start_at is required")
	}

	if req.EndAt == nil {
		return nil, errors.New("end_at is required")
	}

	startDate := req.StartAt.AsTime()
	endDate := req.EndAt.AsTime()

	if startDate.After(endDate) {
		return nil, errors.New("start_at cannot be after end_at")
	}

	summaries, err := s.calculateAccountSummary(ctx, req.AccountIds, startDate, endDate)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate account summary")
	}

	items := make(map[int32]*analyticsv1.GetDebitsAndCreditsSummaryResponse_SummaryItem)

	for accountId, summary := range summaries {
		debitsAmount, _ := summary.TotalDebits.Truncate(0).Float64()
		creditsAmount, _ := summary.TotalCredits.Truncate(0).Float64()

		items[accountId] = &analyticsv1.GetDebitsAndCreditsSummaryResponse_SummaryItem{
			TotalDebitsCount:   summary.DebitsCount,
			TotalCreditsCount:  summary.CreditsCount,
			TotalDebitsAmount:  int32(debitsAmount),
			TotalCreditsAmount: int32(creditsAmount),
		}
	}

	return &analyticsv1.GetDebitsAndCreditsSummaryResponse{
		Items: items,
	}, nil
}

func (s *Service) calculateAccountSummary(
	ctx context.Context,
	accountIds []int32,
	startDate, endDate time.Time,
) (map[int32]*AccountSummary, error) {
	db := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeReadonly))

	type result struct {
		AccountId   int32           `gorm:"column:account_id"`
		IsDebit     bool            `gorm:"column:is_debit"`
		TotalAmount decimal.Decimal `gorm:"column:total_amount"`
		Count       int32           `gorm:"column:count"`
	}

	var results []result
	err := db.Table("double_entries").
		Select("account_id, is_debit, COALESCE(SUM(ABS(amount_in_base_currency)), 0) as total_amount, COUNT(*) as count").
		Where("account_id IN ? AND deleted_at IS NULL", accountIds).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Group("account_id, is_debit").
		Scan(&results).Error

	if err != nil {
		return nil, errors.WithStack(err)
	}

	summaries := make(map[int32]*AccountSummary)

	for _, accountId := range accountIds {
		summaries[accountId] = &AccountSummary{
			TotalDebits:  decimal.Zero,
			TotalCredits: decimal.Zero,
			DebitsCount:  0,
			CreditsCount: 0,
		}
	}

	for _, r := range results {
		sum, ok := summaries[r.AccountId]

		if !ok {
			continue // should not happen
		}

		if r.IsDebit {
			sum.TotalDebits = r.TotalAmount
			sum.DebitsCount = r.Count
		} else {
			sum.TotalCredits = r.TotalAmount
			sum.CreditsCount = r.Count
		}
	}

	return summaries, nil
}
