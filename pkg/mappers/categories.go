package mappers

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (m *Mapper) MapCategory(_ context.Context, category *database.Category) *gomoneypbv1.Category {
	cat := &gomoneypbv1.Category{
		Id:        category.ID,
		Name:      category.Name,
		DeletedAt: nil,
	}

	if category.DeletedAt.Valid {
		cat.DeletedAt = timestamppb.New(category.DeletedAt.Time)
	}

	return cat
}
