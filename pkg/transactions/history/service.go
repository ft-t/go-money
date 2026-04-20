package history

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) Record(ctx context.Context, tx *gorm.DB, req RecordRequest) error {
	snap, err := Snapshot(req.Tx)
	if err != nil {
		return errors.Wrap(err, "snapshot")
	}

	var diff map[string]any
	if req.Previous != nil {
		prevSnap, snapErr := Snapshot(req.Previous)
		if snapErr != nil {
			return errors.Wrap(snapErr, "snapshot prev")
		}
		diff, err = Diff(prevSnap, snap)
		if err != nil {
			return errors.Wrap(err, "diff")
		}
	}

	row := &database.TransactionHistory{
		TransactionID: req.Tx.ID,
		EventType:     req.EventType,
		ActorType:     req.Actor.Type,
		ActorUserID:   req.Actor.UserID,
		ActorRuleID:   req.Actor.RuleID,
		Snapshot:      snap,
		Diff:          diff,
		OccurredAt:    time.Now().UTC(),
	}
	if req.Actor.Detail != "" {
		row.ActorExtra = lo.ToPtr(req.Actor.Detail)
	}
	return errors.WithStack(tx.WithContext(ctx).Create(row).Error)
}

func (s *Service) List(ctx context.Context, transactionID int64) ([]*database.TransactionHistory, error) {
	var rows []*database.TransactionHistory
	if err := database.GetDbWithContext(ctx, database.DbTypeReadonly).
		Where("transaction_id = ?", transactionID).
		Order("occurred_at ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, errors.WithStack(err)
	}
	return rows, nil
}
