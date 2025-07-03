package transactions

import (
	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"context"
	"fmt"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/configuration"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

type Service struct {
	accountCurrencyCache *expirable.LRU[int32, string]
	cfg                  *ServiceConfig
}

func (s *Service) Update(ctx context.Context, msg *transactionsv1.UpdateTransactionRequest) (*transactionsv1.UpdateTransactionResponse, error) {
	//TODO implement me
	panic("implement me")
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

	if len(req.TagIds) > 0 {
		var tagIds []string
		for _, tagId := range req.TagIds {
			tagIds = append(tagIds, fmt.Sprintf("%d", tagId))
		}

		query = query.Where(fmt.Sprintf("tag_ids && Array[%s]", strings.Join(tagIds, ",")))
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

	resp, err := s.CreateBulkInternal(ctx, req, tx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create transactions for request: %v", req)
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return resp, nil
}

func (s *Service) CreateBulkInternal(
	ctx context.Context,
	reqs []*transactionsv1.CreateTransactionRequest,
	tx *gorm.DB,
) ([]*transactionsv1.CreateTransactionResponse, error) {
	//var created []*database.Transaction

	var transactionWithRules []*database.Transaction
	var transactionWithoutRules []*database.Transaction

	for _, req := range reqs {
		if req.TransactionDate == nil {
			return nil, errors.New("transaction date is required")
		}

		newTx := &database.Transaction{
			SourceAmount:            decimal.NullDecimal{},
			SourceCurrency:          "",
			DestinationAmount:       decimal.NullDecimal{},
			DestinationCurrency:     "",
			SourceAccountID:         nil,
			DestinationAccountID:    nil,
			TagIDs:                  req.TagIds,
			CreatedAt:               time.Now().UTC(),
			Notes:                   req.Notes,
			Extra:                   req.Extra,
			TransactionDateTime:     req.TransactionDate.AsTime(),
			TransactionDateOnly:     req.TransactionDate.AsTime(),
			Title:                   req.Title,
			ReferenceNumber:         req.ReferenceNumber,
			InternalReferenceNumber: req.InternalReferenceNumber,
		}

		if newTx.Extra == nil {
			newTx.Extra = map[string]string{}
		}

		var fillRes *fillResponse
		var err error

		switch v := req.GetTransaction().(type) {
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

		if err = tx.Create(newTx).Error; err != nil {
			return nil, errors.WithStack(err)
		}

		if req.SkipRules {
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

	for _, createdTx := range created {
		if err := s.ValidateTransaction(ctx, tx, createdTx); err != nil {
			return nil, errors.Wrapf(err, "failed to validate transaction")
		}
	}

	if err := s.cfg.StatsSvc.HandleTransactions(ctx, tx, created); err != nil {
		return nil, err
	}

	if err := s.cfg.BaseAmountService.RecalculateAmountInBaseCurrency(ctx, tx, created); err != nil {
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

	resp, err := s.CreateBulkInternal(ctx, []*transactionsv1.CreateTransactionRequest{req}, tx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create transaction for request: %v", req)
	}

	if err = tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return resp[0], nil
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

	newTx.TransactionType = gomoneypbv1.TransactionType_TRANSACTION_TYPE_DEPOSIT
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

	newTx.TransactionType = gomoneypbv1.TransactionType_TRANSACTION_TYPE_RECONCILIATION
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
	newTx.TransactionType = gomoneypbv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL

	if req.ForeignCurrency != nil {
		if err = s.ensureCurrencyExists(ctx, *req.ForeignCurrency); err != nil {
			return nil, errors.Wrap(err, "foreign currency does not exist")
		}

		newTx.DestinationCurrency = *req.ForeignCurrency
	}

	if req.ForeignAmount != nil {
		destinationAmount, destinationErr := decimal.NewFromString(*req.ForeignAmount)
		if destinationErr != nil {
			return nil, errors.Wrap(destinationErr, "invalid foreign amount")
		}

		if destinationAmount.IsPositive() || destinationAmount.IsZero() {
			return nil, errors.New("foreign amount must be begative")
		}

		newTx.DestinationAmount = decimal.NewNullDecimal(destinationAmount)

		if newTx.DestinationCurrency == "" {
			return nil, errors.New("foreign currency is required when foreign amount is provided")
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

func (s *Service) ensureCurrencyExists(
	ctx context.Context,
	currency string,
) error {
	return nil // todo
}

func (s *Service) ensureAccountsExistAndCurrencyCorrect(
	_ context.Context,
	dbTx *gorm.DB,
	expectedAccounts map[int32]string,
) (map[int32]*database.Account, error) {
	var accounts []*database.Account

	if err := dbTx.
		Where("id IN ?", lo.Keys(expectedAccounts)).
		Clauses(&clause.Locking{Strength: "UPDATE"}).
		Find(&accounts).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	accCurrencies := map[int32]string{}
	accMap := map[int32]*database.Account{}
	for _, acc := range accounts {
		accMap[acc.ID] = acc
		accCurrencies[acc.ID] = acc.Currency
	}

	for id, expectedCurrency := range expectedAccounts {
		existingCurrency, ok := accCurrencies[id]

		if !ok {
			return nil, errors.Newf("account with id %d not found", id)
		}

		if existingCurrency != expectedCurrency {
			return nil, errors.Newf("account with id %d has currency %s, expected %s", id, existingCurrency, expectedCurrency)
		}
	}

	return accMap, nil
}
