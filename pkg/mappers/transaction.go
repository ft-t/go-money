package mappers

import (
	v1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (m *Mapper) MapTransaction(ctx context.Context, tx *database.Transaction) *v1.Transaction {
	mapped := &v1.Transaction{
		Id:                   tx.ID,
		SourceCurrency:       lo.EmptyableToPtr(tx.SourceCurrency),
		DestinationCurrency:  lo.EmptyableToPtr(tx.DestinationCurrency),
		SourceAccountId:      tx.SourceAccountID,
		DestinationAccountId: tx.DestinationAccountID,
		LabelIds:             tx.LabelIDs,
		CreatedAt:            timestamppb.New(tx.CreatedAt),
		UpdatedAt:            timestamppb.New(tx.UpdatedAt),
		TransactionDate:      timestamppb.New(tx.CreatedAt),
		Type:                 tx.TransactionType,
		Notes:                tx.Notes,
		Extra:                tx.Extra,
		Title:                tx.Title,
	}

	if tx.SourceAmount.Valid {
		mapped.SourceAmount = lo.ToPtr(m.cfg.DecimalSvc.ToString(ctx, tx.SourceAmount.Decimal, tx.SourceCurrency))
	}

	if tx.DestinationAmount.Valid {
		mapped.DestinationAmount = lo.ToPtr(m.cfg.DecimalSvc.ToString(ctx, tx.DestinationAmount.Decimal, tx.DestinationCurrency))
	}

	return mapped
}
