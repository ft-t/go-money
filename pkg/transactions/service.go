package transactions

import (
	"context"
	"fmt"
	"strings"
	"time"

	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	accountCurrencyCache *expirable.LRU[int32, string]
	cfg                  *ServiceConfig
}

type ServiceConfig struct {
	StatsSvc             StatsSvc
	MapperSvc            MapperSvc
	CurrencyConverterSvc CurrencyConverterSvc
	BaseAmountService    BaseAmountSvc
	RuleSvc              RuleSvc
}

func NewService(
	cfg *ServiceConfig,
) *Service {
	return &Service{
		accountCurrencyCache: expirable.NewLRU[int32, string](100, nil, configuration.DefaultCacheTTL),
		cfg:                  cfg,
	}
}

func (s *Service) GetTransactionByIDs(ctx context.Context, ids []int64) ([]*database.Transaction, error) {
	var transactions []*database.Transaction

	if err := database.GetDbWithContext(ctx, database.DbTypeReadonly).
		Where("id IN ?", ids).
		Find(&transactions).
		Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return transactions, nil
}

func (s *Service) List(
	ctx context.Context,
	req *transactionsv1.ListTransactionsRequest,
) (*transactionsv1.ListTransactionsResponse, error) {
	query := database.GetDbWithContext(ctx, database.DbTypeReadonly).Limit(int(req.Limit))

	if req.AmountFrom != nil {
		amountFrom, err := decimal.NewFromString(*req.AmountFrom)
		if err != nil {
			return nil, errors.Wrap(err, "invalid amount_from")
		}

		query = query.Where("(source_account_id is not null and source_amount >= ?) OR (destination_account_id is not null and destination_amount >= ?)",
			amountFrom, amountFrom)
	}

	if req.AmountTo != nil {
		amountTo, err := decimal.NewFromString(*req.AmountTo)
		if err != nil {
			return nil, errors.Wrap(err, "invalid amount_to")
		}

		query = query.Where("(source_account_id is not null and source_amount <= ?) OR (destination_account_id is not null and destination_amount <= ?)",
			amountTo, amountTo)
	}

	if req.FromDate != nil {
		query = query.Where("transaction_date_time >= ?", req.FromDate.AsTime())
	}

	if req.ToDate != nil {
		query = query.Where("transaction_date_time <= ?", req.ToDate.AsTime())
	}

	if req.TextQuery != nil {
		query = query.Where("title ILIKE ?", "%"+*req.TextQuery+"%")
	}

	if len(req.DestinationAccountIds) > 0 {
		query = query.Where("destination_account_id IN ?", req.DestinationAccountIds)
	}

	if len(req.SourceAccountIds) > 0 {
		query = query.Where("source_account_id IN ?", req.SourceAccountIds)
	}

	if len(req.AnyAccountIds) > 0 {
		query = query.Where("source_account_id IN ? OR destination_account_id IN ?", req.AnyAccountIds, req.AnyAccountIds)
	}

	if len(req.TransactionTypes) > 0 {
		query = query.Where("transaction_type IN ?", lo.Map(req.TransactionTypes, func(t gomoneypbv1.TransactionType, _ int) int32 {
			return int32(t)
		}))
	}

	if len(req.CategoryIds) > 0 {
		query = query.Where("category_id IN ?", req.CategoryIds)
	}

	if len(req.TagIds) > 0 {
		var tagIds []string
		for _, tagId := range req.TagIds {
			tagIds = append(tagIds, fmt.Sprintf("%d", tagId))
		}

		query = query.Where(fmt.Sprintf("tag_ids && Array[%s]", strings.Join(tagIds, ",")))
	}

	if len(req.Ids) > 0 {
		query = query.Where("id IN ?", req.Ids)
	}

	for _, sort := range req.Sort {
		switch sort.Field {
		default:
			query = query.Order(clause.OrderByColumn{
				Column: clause.Column{
					Table: "transactions",
					Name:  "transaction_date_time",
				},
				Desc: !sort.Ascending,
			})
		}
	}

	var transactions []*database.Transaction

	var count int64
	if err := query.Model(transactions).Count(&count).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := query.Limit(int(req.Limit)).Offset(int(req.Skip)).Find(&transactions).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	resp := &transactionsv1.ListTransactionsResponse{
		Transactions: nil,
		TotalCount:   count,
	}

	for _, tx := range transactions {
		resp.Transactions = append(resp.Transactions, s.cfg.MapperSvc.MapTransaction(ctx, tx))
	}

	return resp, nil
}

func (s *Service) CreateBulk(
	ctx context.Context,
	req []*transactionsv1.CreateTransactionRequest,
) ([]*transactionsv1.CreateTransactionResponse, error) {
	tx := database.GetDbWithContext(ctx, database.DbTypeMaster).Begin()
	defer tx.Rollback()
	ctx = database.WithContext(ctx, tx)

	var bulkRequests []*BulkRequest
	for _, r := range req {
		bulkRequests = append(bulkRequests, &BulkRequest{
			Req: r,
		})
	}

	resp, err := s.CreateBulkInternal(ctx, bulkRequests, tx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create transactions for request: %v", req)
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return resp, nil
}

type BulkRequest struct {
	OriginalTx *database.Transaction // for update
	Req        *transactionsv1.CreateTransactionRequest
}

func (s *Service) CreateBulkInternal(
	ctx context.Context,
	reqs []*BulkRequest,
	tx *gorm.DB,
) ([]*transactionsv1.CreateTransactionResponse, error) {
	var transactionWithRules []*database.Transaction
	var transactionWithoutRules []*database.Transaction

	var originalTxs []*database.Transaction

	for _, req := range reqs {
		if req.Req.TransactionDate == nil {
			return nil, errors.New("transaction date is required")
		}

		if req.OriginalTx != nil { // save list of original transactions for update
			originalTxs = append(originalTxs, req.OriginalTx)
		}

		newTx := &database.Transaction{
			SourceAmount:            decimal.NullDecimal{},
			SourceCurrency:          "",
			DestinationAmount:       decimal.NullDecimal{},
			DestinationCurrency:     "",
			SourceAccountID:         nil,
			DestinationAccountID:    nil,
			TagIDs:                  req.Req.TagIds,
			CreatedAt:               time.Now().UTC(),
			Notes:                   req.Req.Notes,
			Extra:                   req.Req.Extra,
			TransactionDateTime:     req.Req.TransactionDate.AsTime(),
			TransactionDateOnly:     req.Req.TransactionDate.AsTime(),
			Title:                   req.Req.Title,
			ReferenceNumber:         req.Req.ReferenceNumber,
			InternalReferenceNumber: req.Req.InternalReferenceNumber,
			CategoryID:              req.Req.CategoryId,
		}

		if req.OriginalTx != nil {
			newTx.ID = req.OriginalTx.ID
			newTx.CreatedAt = req.OriginalTx.CreatedAt
			newTx.UpdatedAt = time.Now().UTC()

			req.OriginalTx = newTx // swap
		}

		if newTx.Extra == nil {
			newTx.Extra = map[string]string{}
		}

		var fillRes *fillResponse
		var err error

		switch v := req.Req.GetTransaction().(type) {
		case *transactionsv1.CreateTransactionRequest_TransferBetweenAccounts:
			if fillRes, err = s.fillTransferBetweenAccounts(ctx, tx, v.TransferBetweenAccounts, newTx); err != nil {
				return nil, err
			}
		case *transactionsv1.CreateTransactionRequest_Withdrawal:
			if fillRes, err = s.fillWithdrawal(ctx, v.Withdrawal, newTx); err != nil {
				return nil, err
			}
		case *transactionsv1.CreateTransactionRequest_Deposit:
			if fillRes, err = s.fillDeposit(ctx, v.Deposit, newTx); err != nil {
				return nil, err
			}
		case *transactionsv1.CreateTransactionRequest_Reconciliation:
			if fillRes, err = s.fillReconciliation(ctx, v.Reconciliation, newTx); err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("invalid transaction type")
		}

		for _, acc := range fillRes.Accounts {
			if acc.FirstTransactionAt == nil || acc.FirstTransactionAt.After(newTx.TransactionDateTime) {
				acc.FirstTransactionAt = &newTx.TransactionDateTime
			}
		}

		// validate wallet transaction date

		if req.OriginalTx == nil {
			if err = tx.Create(newTx).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			if err = tx.Updates(newTx).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		}

		if req.Req.SkipRules {
			transactionWithoutRules = append(transactionWithoutRules, newTx)
		} else {
			transactionWithRules = append(transactionWithRules, newTx)
		}
	}

	if len(transactionWithRules) > 0 {
		modifiedTxs, err := s.cfg.RuleSvc.ProcessTransactions(ctx, transactionWithRules) // run rule engine can change transactions
		if err != nil {
			return nil, errors.Wrap(err, "failed to process transactions with rules")
		}

		transactionWithRules = modifiedTxs
	}

	created := append(transactionWithRules, transactionWithoutRules...)

	return s.FinalizeTransactions(ctx, tx, created, originalTxs)
}

func (s *Service) CreateRawTransaction(
	ctx context.Context,
	newTx *database.Transaction,
) (*transactionsv1.CreateTransactionResponse, error) {
	tx := database.GetDbWithContext(ctx, database.DbTypeMaster).Begin()
	defer tx.Rollback()
	ctx = database.WithContext(ctx, tx)

	if newTx.Extra == nil {
		newTx.Extra = map[string]string{}
	}

	if err := tx.Create(newTx).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to create transaction: %v", newTx)
	}

	resp, err := s.FinalizeTransactions(ctx, tx, []*database.Transaction{newTx}, nil)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return resp[0], nil
}

func (s *Service) FinalizeTransactions(
	ctx context.Context,
	tx *gorm.DB,
	created []*database.Transaction,
	originalTxs []*database.Transaction,
) ([]*transactionsv1.CreateTransactionResponse, error) {
	// todo validate

	// include original as we need to ensure previous history is correct now
	if err = s.cfg.StatsSvc.HandleTransactions(ctx, tx, append(created, originalTxs...)); err != nil {
		return nil, err
	}

	if err = s.cfg.BaseAmountService.RecalculateAmountInBaseCurrency(ctx, tx, created); err != nil {
		return nil, errors.Wrap(err, "failed to recalculate amounts in base currency")
	}

	var finalRes []*transactionsv1.CreateTransactionResponse

	for _, createdTx := range created {
		finalRes = append(finalRes, &transactionsv1.CreateTransactionResponse{
			Transaction: s.cfg.MapperSvc.MapTransaction(ctx, createdTx),
		})
	}

	return finalRes, nil
}

func (s *Service) Create(
	ctx context.Context,
	req *transactionsv1.CreateTransactionRequest,
) (*transactionsv1.CreateTransactionResponse, error) {
	tx := database.GetDbWithContext(ctx, database.DbTypeMaster).Begin()
	defer tx.Rollback()
	ctx = database.WithContext(ctx, tx)

	resp, err := s.CreateBulkInternal(ctx, []*BulkRequest{
		{
			Req: req,
		},
	}, tx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create transaction for request: %v", req)
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return resp[0], nil
}

func (s *Service) Update(
	ctx context.Context,
	req *transactionsv1.UpdateTransactionRequest,
) (*transactionsv1.UpdateTransactionResponse, error) {
	tx := database.GetDbWithContext(ctx, database.DbTypeMaster).Begin()
	defer tx.Rollback()
	ctx = database.WithContext(ctx, tx)

	var existingTx database.Transaction
	if err := tx.Where("id = ?", req.Id).
		First(&existingTx).Error; err != nil {
		return nil, errors.Wrap(err, "failed to find existing transaction")
	}

	resp, err := s.CreateBulkInternal(ctx, []*BulkRequest{
		{
			Req:        req.Transaction,
			OriginalTx: &existingTx, // we need to update existing transaction
		},
	}, tx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create transaction for request: %v", req)
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &transactionsv1.UpdateTransactionResponse{
		Transaction: resp[0].Transaction,
	}, nil
}

func (s *Service) fillDeposit(
	_ context.Context,
	req *transactionsv1.Deposit,
	newTx *database.Transaction,
) (*fillResponse, error) {
	destinationAmount, err := decimal.NewFromString(req.DestinationAmount)
	if err != nil {
		return nil, errors.Wrap(err, "invalid destination amount")
	}

	newTx.TransactionType = gomoneypbv1.TransactionType_TRANSACTION_TYPE_INCOME
	newTx.DestinationAmount = decimal.NewNullDecimal(destinationAmount)
	newTx.DestinationCurrency = req.DestinationCurrency
	newTx.DestinationAccountID = &req.DestinationAccountId

	return &fillResponse{}, nil
}

func (s *Service) fillReconciliation(
	_ context.Context,
	req *transactionsv1.Reconciliation,
	newTx *database.Transaction,
) (*fillResponse, error) {
	destinationAmount, err := decimal.NewFromString(req.DestinationAmount)
	if err != nil {
		return nil, errors.Wrap(err, "invalid destination amount")
	}

	newTx.TransactionType = gomoneypbv1.TransactionType_TRANSACTION_TYPE_ADJUSTMENT
	newTx.DestinationAmount = decimal.NewNullDecimal(destinationAmount)
	newTx.DestinationCurrency = req.DestinationCurrency
	newTx.DestinationAccountID = &req.DestinationAccountId

	return &fillResponse{}, nil
}

func (s *Service) fillWithdrawal(
	ctx context.Context,
	req *transactionsv1.Withdrawal,
	newTx *database.Transaction,
) (*fillResponse, error) {
	sourceAmount, err := decimal.NewFromString(req.SourceAmount)
	if err != nil {
		return nil, errors.Wrap(err, "invalid source amount")
	}

	newTx.SourceAmount = decimal.NewNullDecimal(sourceAmount)
	newTx.SourceCurrency = req.SourceCurrency
	newTx.SourceAccountID = &req.SourceAccountId
	newTx.TransactionType = gomoneypbv1.TransactionType_TRANSACTION_TYPE_EXPENSE

	// fx
	if req.FxSourceCurrency != nil {
		if err = s.ensureCurrencyExists(ctx, *req.FxSourceCurrency); err != nil {
			return nil, errors.Wrap(err, "foreign currency does not exist")
		}

		newTx.FxSourceCurrency = *req.FxSourceCurrency
	}

	if req.FxSourceAmount != nil {
		fxAmount, destinationErr := decimal.NewFromString(*req.FxSourceAmount)
		if destinationErr != nil {
			return nil, errors.Wrap(destinationErr, "invalid foreign amount")
		}

		if fxAmount.IsPositive() || fxAmount.IsZero() {
			return nil, errors.New("foreign amount must be negative")
		}

		newTx.FxSourceAmount = decimal.NewNullDecimal(fxAmount)

		if newTx.FxSourceCurrency == "" {
			return nil, errors.New("foreign currency is required when foreign amount is provided")
		}
	}

	// dest
	if req.DestinationCurrency != nil {
		if err = s.ensureCurrencyExists(ctx, *req.DestinationCurrency); err != nil {
			return nil, errors.Wrap(err, "destination currency does not exist")
		}

		newTx.DestinationCurrency = *req.DestinationCurrency
	}

	if req.DestinationAmount != nil {
		destinationAmount, destinationErr := decimal.NewFromString(*req.DestinationAmount)
		if destinationErr != nil {
			return nil, errors.Wrap(destinationErr, "invalid destination amount")
		}

		if destinationAmount.IsNegative() || destinationAmount.IsZero() {
			return nil, errors.New("destination amount must be positive")
		}

		newTx.DestinationAmount = decimal.NewNullDecimal(destinationAmount)
		if newTx.DestinationCurrency == "" {
			return nil, errors.New("destination currency is required when destination amount is provided")
		}
	}

	return &fillResponse{}, nil
}

func (s *Service) fillTransferBetweenAccounts(
	ctx context.Context,
	dbTx *gorm.DB,
	req *transactionsv1.TransferBetweenAccounts,
	newTx *database.Transaction,
) (*fillResponse, error) {
	if req.SourceAccountId <= 0 {
		return nil, errors.New("source account id is required")
	}

	if req.DestinationAccountId <= 0 {
		return nil, errors.New("destination account id is required")
	}

	accounts, err := s.ensureAccountsExistAndCurrencyCorrect(ctx, dbTx, map[int32]string{
		req.SourceAccountId:      req.SourceCurrency,
		req.DestinationAccountId: req.DestinationCurrency,
	})
	if err != nil {
		return nil, err
	}

	sourceAmount, err := decimal.NewFromString(req.SourceAmount)
	if err != nil {
		return nil, errors.Wrap(err, "invalid source amount")
	}

	if sourceAmount.IsPositive() || sourceAmount.IsZero() {
		return nil, errors.New("source amount must be negative")
	}

	destinationAmount, err := decimal.NewFromString(req.DestinationAmount)
	if err != nil {
		return nil, errors.Wrap(err, "invalid destination amount")
	}

	if destinationAmount.IsNegative() || destinationAmount.IsZero() {
		return nil, errors.New("destination amount must be positive")
	}

	newTx.TransactionType = gomoneypbv1.TransactionType_TRANSACTION_TYPE_TRANSFER_BETWEEN_ACCOUNTS

	newTx.SourceAmount = decimal.NewNullDecimal(sourceAmount)
	newTx.DestinationAmount = decimal.NewNullDecimal(destinationAmount)

	newTx.SourceAccountID = &req.SourceAccountId
	newTx.DestinationAccountID = &req.DestinationAccountId

	newTx.SourceCurrency = req.SourceCurrency
	newTx.DestinationCurrency = req.DestinationCurrency

	return &fillResponse{
		Accounts: accounts,
	}, nil
}
