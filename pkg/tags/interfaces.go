package tags

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package tags_test -source=interfaces.go

type Mapper interface {
	MapTag(ctx context.Context, tag *database.Tag) *gomoneypbv1.Tag
}
