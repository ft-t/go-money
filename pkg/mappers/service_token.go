package mappers

import (
	"context"

	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/ft-t/go-money/pkg/database"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (m *Mapper) MapServiceToken(_ context.Context, token *database.ServiceToken) *gomoneypbv1.ServiceToken {
	res := &gomoneypbv1.ServiceToken{
		Id:        token.ID,
		Name:      token.Name,
		CreatedAt: timestamppb.New(token.CreatedAt),
		ExpiresAt: timestamppb.New(token.ExpiresAt),
	}

	if token.DeletedAt.Valid {
		res.DeletedAt = timestamppb.New(token.DeletedAt.Time)
	}

	return res
}
