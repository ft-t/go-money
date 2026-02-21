package importers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/ft-t/go-money/pkg/transactions"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
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

func (i *Importer) CheckDuplicates(
	ctx context.Context,
	requests []*transactionsv1.CreateTransactionRequest,
	skipDuplicateRefCheck bool,
) ([]*DeduplicationItem, error) {
	var allRefs []string
	refToItem := map[string]*DeduplicationItem{}
	var items []*DeduplicationItem

	for _, req := range requests {
		var validRefs []string
		for _, ref := range req.InternalReferenceNumbers {
			trimmed := strings.TrimSpace(ref)
			if trimmed != "" {
				validRefs = append(validRefs, trimmed)
			}
		}

		if len(validRefs) == 0 {
			return nil, errors.New("all transactions must have at least one reference number for deduplication")
		}

		for idx, ref := range validRefs {
			if _, exists := refToItem[ref]; exists {
				if !skipDuplicateRefCheck {
					return nil, errors.Errorf("duplicate reference number found in import data: ref=%s. raw=%v", ref,
						refToItem[ref].CreateRequest.Notes)
				}

				counter := 1
				newRef := fmt.Sprintf("%s_%d", ref, counter)
				for _, exists := refToItem[newRef]; exists; _, exists = refToItem[newRef] {
					counter++
					newRef = fmt.Sprintf("%s_%d", ref, counter)
				}

				validRefs[idx] = newRef
			}
		}

		req.InternalReferenceNumbers = validRefs

		item := &DeduplicationItem{CreateRequest: req}
		items = append(items, item)

		for _, ref := range validRefs {
			refToItem[ref] = item
		}

		allRefs = append(allRefs, validRefs...)
	}

	for _, chunk := range lo.Chunk(allRefs, boilerplate.DefaultBatchSize) {
		var existingTransactions []*struct {
			InternalReferenceNumbers pq.StringArray `gorm:"type:text[]"`
			ID                       int64
		}

		if err := database.FromContext(ctx, database.GetDb(database.DbTypeMaster)).
			Model(&database.Transaction{}).
			Where("deleted_at is null").
			Where("internal_reference_numbers && ?", pq.Array(chunk)).
			Select("internal_reference_numbers, id").Find(&existingTransactions).Error; err != nil {
			return nil, errors.Wrap(err, "failed to check existing transactions")
		}

		for _, record := range existingTransactions {
			for _, ref := range record.InternalReferenceNumbers {
				if item, exists := refToItem[ref]; exists {
					item.DuplicationTransactionID = &record.ID
				}
			}
		}
	}

	return items, nil
}

func (i *Importer) Import(
	ctx context.Context,
	req *importv1.ImportTransactionsRequest,
) (*importv1.ImportTransactionsResponse, error) {
	parsed, err := i.ParseInternal(ctx, &importv1.ParseTransactionsRequest{
		Content:                     req.Content,
		Source:                      req.Source,
		TreatDatesAsUtc:             req.TreatDatesAsUtc,
		SkipDuplicateReferenceCheck: req.SkipDuplicateReferenceCheck,
	})
	if err != nil {
		return nil, err
	}

	var allTransactions []*transactions.BulkRequest
	var duplicateCount int

	for _, item := range parsed {
		if item.DuplicationTransactionID != nil {
			duplicateCount += 1
			continue
		}

		allTransactions = append(allTransactions, &transactions.BulkRequest{
			Req: item.CreateRequest,
		})
	}

	tx := database.FromContext(ctx, database.GetDb(database.DbTypeMaster)).Begin()
	defer tx.Rollback()
	ctx = database.WithContext(ctx, tx)

	transactionResp, transactionErr := i.cfg.TransactionSvc.CreateBulkInternal(
		ctx,
		allTransactions,
		tx,
		transactions.UpsertOptions{
			SkipAccountSourceDestValidation: true,
		},
	)

	if transactionErr != nil {
		return nil, errors.Wrap(transactionErr, "failed to create transactions")
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.Wrap(err, "failed to commit transaction")
	}

	return &importv1.ImportTransactionsResponse{
		ImportedCount:  int32(len(transactionResp)),
		DuplicateCount: int32(duplicateCount),
	}, nil
}

func (i *Importer) Parse(
	ctx context.Context,
	req *importv1.ParseTransactionsRequest,
) (*importv1.ParseTransactionsResponse, error) {
	parsed, err := i.ParseInternal(ctx, req)
	if err != nil {
		return nil, err
	}

	converted, err := i.ConvertRequestsToTransactions(ctx, parsed)
	if err != nil {
		return nil, err
	}

	return &importv1.ParseTransactionsResponse{
		Transactions: converted,
	}, nil
}

func (i *Importer) ParseInternal(
	ctx context.Context,
	req *importv1.ParseTransactionsRequest,
) ([]*DeduplicationItem, error) {
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
			TreatDatesAsUtc: req.TreatDatesAsUtc,
		},
	})
	if err != nil {
		log.Error().
			Err(err).
			Str("source", req.Source.String()).
			Int("accounts_count", len(accounts)).
			Int("tags_count", len(tags)).
			Int("categories_count", len(categories)).
			Msg("import parse failed")

		return nil, errors.Wrap(err, "parse failed")
	}

	if len(parsed.CreateRequests) == 0 {
		return nil, errors.New("no transactions found in import data")
	}

	batchID := uuid.NewString()

	for _, r := range parsed.CreateRequests {
		if r.Extra == nil {
			r.Extra = map[string]string{}
		}

		r.Extra["import_batch_id"] = batchID
	}

	items, err := i.CheckDuplicates(ctx, parsed.CreateRequests, req.SkipDuplicateReferenceCheck)
	if err != nil {
		log.Error().
			Err(err).
			Str("source", req.Source.String()).
			Int("transactions_count", len(parsed.CreateRequests)).
			Msg("duplicate check failed")

		return nil, errors.Wrap(err, "failed to check for duplicate transactions")
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreateRequest.TransactionDate.AsTime().Before(items[j].CreateRequest.TransactionDate.AsTime())
	})

	return items, nil
}

func (i *Importer) ConvertRequestsToTransactions(
	ctx context.Context,
	requests []*DeduplicationItem,
) ([]*importv1.ParseTransactionsResponse_ParsedTransaction, error) {
	var result []*importv1.ParseTransactionsResponse_ParsedTransaction

	for _, req := range requests {
		if !req.CreateRequest.HasTransaction() { // it means parsing failed, but we want to show raw transaction to user, so user can decide what to do
			result = append(result, &importv1.ParseTransactionsResponse_ParsedTransaction{
				DuplicateTransactionId: req.DuplicationTransactionID,
				Transaction: i.cfg.MapperSvc.MapTransaction(ctx, &database.Transaction{
					Title:           req.CreateRequest.Title,
					Notes:           req.CreateRequest.Notes,
					TransactionType: gomoneypbv1.TransactionType_TRANSACTION_TYPE_UNSPECIFIED,
				}),
			})

			continue
		}

		converted, err := i.cfg.TransactionSvc.ConvertRequestToTransaction(ctx, req.CreateRequest, nil)
		if err != nil {
			log.Error().
				Err(err).
				Str("title", req.CreateRequest.Title).
				Str("notes", req.CreateRequest.Notes).
				Msg("failed to convert request to transaction")

			return nil, errors.Wrap(err, "failed to convert request to transaction")
		}

		result = append(result, &importv1.ParseTransactionsResponse_ParsedTransaction{
			Transaction:            i.cfg.MapperSvc.MapTransaction(ctx, converted),
			DuplicateTransactionId: req.DuplicationTransactionID,
		})
	}

	return result, nil
}
