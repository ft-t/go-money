package mappers

import (
	v1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Mapper struct {
	cfg *MapperConfig
}

type MapperConfig struct {
	DecimalSvc DecimalSvc
}

func NewMapper(
	cfg *MapperConfig,
) *Mapper {
	return &Mapper{
		cfg: cfg,
	}
}

func (m *Mapper) MapAccount(ctx context.Context, acc *database.Account) *v1.Account {
	mapped := &v1.Account{
		Id:               acc.ID,
		Name:             acc.Name,
		Currency:         acc.Currency,
		CurrentBalance:   m.cfg.DecimalSvc.ToString(ctx, acc.CurrentBalance, acc.Currency),
		Extra:            acc.Extra,
		UpdatedAt:        timestamppb.New(acc.LastUpdatedAt),
		DeletedAt:        nil,
		Type:             acc.Type,
		Note:             acc.Note,
		LiabilityPercent: nil,
		Iban:             acc.Iban,
		AccountNumber:    acc.AccountNumber,
		DisplayOrder:     acc.DisplayOrder,
	}

	if acc.LiabilityPercent.Valid {
		mapped.LiabilityPercent = lo.ToPtr(acc.LiabilityPercent.Decimal.String())
	}

	if acc.DeletedAt.Valid {
		mapped.DeletedAt = timestamppb.New(acc.DeletedAt.Time)
	}

	return mapped
}
