package importers

import (
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"context"
)

type Importer struct {
	impl map[importv1.ImportSource]Implementation
}

func NewImporter(
	impl ...Implementation,
) *Importer {
	implementations := make(map[importv1.ImportSource]Implementation)
	for _, i := range impl {
		implementations[i.Type()] = i
	}

	return &Importer{
		impl: implementations,
	}
}

func (i *Importer) Import(
	ctx context.Context,
	req *importv1.ImportTransactionsRequest,
) (*importv1.ImportTransactionsResponse, error) {
	//TODO implement me
	panic("implement me")
}
