package mappers

import (
	"context"

	v1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (m *Mapper) MapTransaction(ctx context.Context, tx *database.Transaction) *v1.Transaction {
	mapped := &v1.Transaction{
		Id:                      tx.ID,
		SourceCurrency:          lo.EmptyableToPtr(tx.SourceCurrency),
		DestinationCurrency:     lo.EmptyableToPtr(tx.DestinationCurrency),
		SourceAccountId:         tx.SourceAccountID,
		DestinationAccountId:    tx.DestinationAccountID,
		TagIds:                  tx.TagIDs,
		CreatedAt:               timestamppb.New(tx.CreatedAt),
		UpdatedAt:               timestamppb.New(tx.UpdatedAt),
		TransactionDate:         timestamppb.New(tx.TransactionDateTime),
		Type:                    tx.TransactionType,
		Notes:                   tx.Notes,
		Extra:                   tx.Extra,
		Title:                   tx.Title,
		InternalReferenceNumber: tx.InternalReferenceNumber,
		ReferenceNumber:         tx.ReferenceNumber,
		CategoryId:              tx.CategoryID,
		FxSourceCurrency:        lo.EmptyableToPtr(tx.FxSourceCurrency),
	}

	if tx.SourceAmount.Valid {
		mapped.SourceAmount = lo.ToPtr(m.cfg.DecimalSvc.ToString(ctx, tx.SourceAmount.Decimal, tx.SourceCurrency))
	}

	if tx.DestinationAmount.Valid {
		mapped.DestinationAmount = lo.ToPtr(m.cfg.DecimalSvc.ToString(ctx, tx.DestinationAmount.Decimal, tx.DestinationCurrency))
	}

	if tx.FxSourceAmount.Valid {
		mapped.FxSourceAmount = lo.ToPtr(m.cfg.DecimalSvc.ToString(ctx, tx.FxSourceAmount.Decimal, tx.FxSourceCurrency))
	}

	return mapped
}
