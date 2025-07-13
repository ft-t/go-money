package categories

import (
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"github.com/ft-t/go-money/pkg/database"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package categories_test -source=interfaces.go
type Mapper interface {
	MapCategory(ctx context.Context, category *database.Category) *gomoneypbv1.Category
}
