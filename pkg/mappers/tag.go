package mappers

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
)

func (m *Mapper) MapTag(_ context.Context, tag *database.Tag) *gomoneypbv1.Tag {
	return &gomoneypbv1.Tag{
		Id:    tag.ID,
		Name:  tag.Name,
		Color: tag.Color,
		Icon:  tag.Icon,
	}
}
