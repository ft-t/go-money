package importers

import (
	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"context"
	"encoding/base64"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
)

type Importer struct {
	impl          map[importv1.ImportSource]Implementation
	accountSvc    AccountSvc
	tagSvc        TagSvc
	categoriesSvc CategoriesSvc
}

func NewImporter(
	accountSvc AccountSvc,
	tagSvc TagSvc,
	categoriesSvc CategoriesSvc,
	impl ...Implementation,
) *Importer {
	implementations := make(map[importv1.ImportSource]Implementation)
	for _, i := range impl {
		implementations[i.Type()] = i
	}

	return &Importer{
		impl:          implementations,
		accountSvc:    accountSvc,
		tagSvc:        tagSvc,
		categoriesSvc: categoriesSvc,
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

	tags, err := i.tagSvc.GetAllTags(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tags")
	}

	tagMap := make(map[string]*database.Tag)
	for _, tag := range tags {
		tagMap[tag.Name] = tag
	}

	categories, err := i.categoriesSvc.GetAllCategories(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get categories")
	}

	categoryMap := make(map[string]*database.Category)
	for _, category := range categories {
		categoryMap[category.Name] = category
	}

	resp, err := impl.Import(ctx, &ImportRequest{
		Data:            decoded,
		Accounts:        accounts,
		Tags:            tagMap,
		SkipRules:       req.SkipRules,
		TreatDatesAsUtc: req.TreatDatesAsUtc,
		Categories:      categoryMap,
	})
	if err != nil {
		return nil, errors.Wrap(err, "import failed")
	}

	return resp, nil
}
