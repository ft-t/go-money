package importers

import (
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"context"
)

type Implementation interface {
	Import(ctx context.Context, req *importv1.ImportTransactionsRequest) (*importv1.ImportTransactionsResponse, error)
	Type() importv1.ImportSource
}
