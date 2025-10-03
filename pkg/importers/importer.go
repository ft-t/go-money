package importers

import (
	"context"
	"sort"
	"time"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/samber/lo"
)

type Importer struct {
	cfg             *ImporterConfig
	implementations map[importv1.ImportSource]Implementation
}

type ImporterConfig struct {
	AccountSvc     AccountSvc
	TagSvc         TagSvc
	CategoriesSvc  CategoriesSvc
	TransactionSvc TransactionSvc
	MapperSvc      MapperSvc
}

func NewImporter(
	cfg *ImporterConfig,
	impl ...Implementation,
) *Importer {
	implementations := make(map[importv1.ImportSource]Implementation)
	for _, i := range impl {
		implementations[i.Type()] = i
	}

	return &Importer{
		implementations: implementations,
		cfg:             cfg,
	}
}

func (i *Importer) Import(
	ctx context.Context,
	req *importv1.ImportTransactionsRequest,
) (*importv1.ImportTransactionsResponse, error) {
	parsed, err := i.ParseInternal(ctx, &importv1.ParseTransactionsRequest{
		Content:         req.FileContent,
		Source:          req.Source,
		TreatDatesAsUtc: req.TreatDatesAsUtc,
	})
	if err != nil {
		return nil, err
	}

	//
	//journalIDs := lo.Keys(newTxs)
	//duplicateCount := 0
	//
	//for _, chunk := range lo.Chunk(journalIDs, boilerplate.DefaultBatchSize) {
	//	var existingRecords []string
	//
	//	if err = database.FromContext(ctx, database.GetDb(database.DbTypeMaster)).
	//		Model(&database.ImportDeduplication{}).
	//		Where("import_source = ?", f.Type().Number()).
	//		Where("key in ?", chunk).
	//		Pluck("key", &existingRecords).Error; err != nil {
	//		return nil, errors.Wrap(err, "failed to check existing transactions")
	//	}
	//
	//	for _, record := range existingRecords {
	//		delete(newTxs, record)
	//
	//		duplicateCount += 1
	//	}
	//}
	//
	//if len(newTxs) == 0 {
	//	return &importv1.ImportTransactionsResponse{
	//		ImportedCount:  0,
	//		DuplicateCount: int32(duplicateCount),
	//	}, nil
	//}
	//
	//var allTransactions []*transactions.BulkRequest
	//for _, tx := range newTxs {
	//	allTransactions = append(allTransactions, &transactions.BulkRequest{
	//		Req: tx,
	//	})
	//}
	//
	//sort.Slice(allTransactions, func(i, j int) bool {
	//	return allTransactions[i].Req.TransactionDate.AsTime().Before(allTransactions[j].Req.TransactionDate.AsTime())
	//})
	//
	//tx := database.FromContext(ctx, database.GetDb(database.DbTypeMaster)).Begin()
	//defer tx.Rollback()
	//ctx = database.WithContext(ctx, tx)
	//
	//transactionResp, transactionErr := f.transactionService.CreateBulkInternal(
	//	ctx,
	//	allTransactions,
	//	tx,
	//	transactions.UpsertOptions{
	//		SkipAccountSourceDestValidation: true, // todo from request
	//	},
	//)
	//if transactionErr != nil {
	//	return nil, errors.Wrap(transactionErr, "failed to create transactions")
	//}
	//
	//var deduplicationRecords []*database.ImportDeduplication
	//for _, record := range transactionResp {
	//	deduplicationRecords = append(deduplicationRecords, &database.ImportDeduplication{
	//		ImportSource:  importv1.ImportSource_IMPORT_SOURCE_FIREFLY,
	//		Key:           *record.Transaction.InternalReferenceNumber,
	//		CreatedAt:     time.Now(),
	//		TransactionID: record.Transaction.Id,
	//	})
	//}
	//
	//if err = tx.CreateInBatches(&deduplicationRecords, boilerplate.DefaultBatchSize).Error; err != nil {
	//	return nil, errors.Wrap(err, "failed to create deduplication records")
	//}
	//
	//if err = tx.Commit().Error; err != nil {
	//	return nil, errors.Wrap(err, "failed to commit transaction")
	//}
	//
	//return &importv1.ImportTransactionsResponse{
	//	ImportedCount:  int32(len(allTransactions)),
	//	DuplicateCount: int32(duplicateCount),
	//}, nil
	//
	return nil, nil
}

func (i *Importer) Parse(
	ctx context.Context,
	req *importv1.ParseTransactionsRequest,
) (*importv1.ParseTransactionsResponse, error) {
	parsed, err := i.ParseInternal(ctx, req)
	if err != nil {
		return nil, err
	}

	converted, err := i.ConvertRequestsToTransactions(ctx, parsed.CreateRequests)

	return &importv1.ParseTransactionsResponse{
		Transactions: converted,
	}, nil
}

func (i *Importer) ParseInternal(
	ctx context.Context,
	req *importv1.ParseTransactionsRequest,
) (*ParseResponse, error) {
	impl, ok := i.implementations[req.Source]
	if !ok {
		return nil, errors.New("unsupported import source")
	}

	accounts, err := i.cfg.AccountSvc.GetAllAccounts(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get accounts")
	}

	tags, err := i.cfg.TagSvc.GetAllTags(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get tags")
	}

	tagMap := make(map[string]*database.Tag)
	for _, tag := range tags {
		tagMap[tag.Name] = tag
	}

	categories, err := i.cfg.CategoriesSvc.GetAllCategories(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get categories")
	}

	categoryMap := make(map[string]*database.Category)
	for _, category := range categories {
		categoryMap[category.Name] = category
	}

	parsed, err := impl.Parse(ctx, &ParseRequest{
		ImportRequest: ImportRequest{
			Data:            req.Content,
			Accounts:        accounts,
			Tags:            tagMap,
			Categories:      categoryMap,
			SkipRules:       false, // todo
			TreatDatesAsUtc: false, // todo
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "parse failed")
	}

	return parsed, nil
}

func (i *Importer) ConvertRequestsToTransactions(
	ctx context.Context,
	requests []*transactionsv1.CreateTransactionRequest,
) ([]*gomoneypbv1.Transaction, error) {
	var result []*gomoneypbv1.Transaction

	for _, req := range requests {
		converted, err := i.cfg.TransactionSvc.ConvertRequestToTransaction(ctx, req, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert request to transaction")
		}

		result = append(result, i.cfg.MapperSvc.MapTransaction(ctx, converted))
	}

	return result, nil
}
