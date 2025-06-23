package importers

import (
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"context"
	"encoding/base64"
	"github.com/cockroachdb/errors"
)

type Importer struct {
	impl       map[importv1.ImportSource]Implementation
	accountSvc AccountSvc
}

func NewImporter(
	accountSvc AccountSvc,
	impl ...Implementation,
) *Importer {
	implementations := make(map[importv1.ImportSource]Implementation)
	for _, i := range impl {
		implementations[i.Type()] = i
	}

	return &Importer{
		impl:       implementations,
		accountSvc: accountSvc,
	}
}

func (i *Importer) Import(
	ctx context.Context,
	req *importv1.ImportTransactionsRequest,
) (*importv1.ImportTransactionsResponse, error) {
	impl, ok := i.impl[req.Source]
	if !ok {
		return nil, errors.New("unsupported import source")
	}

	decoded, err := base64.StdEncoding.DecodeString(req.FileContent)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode file content")
	}

	accounts, err := i.accountSvc.GetAllAccounts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get accounts")
	}

	resp, err := impl.Import(ctx, &ImportRequest{
		Data:     decoded,
		Accounts: accounts,
	})
	if err != nil {
		return nil, errors.Wrap(err, "import failed")
	}

	return resp, nil
}
