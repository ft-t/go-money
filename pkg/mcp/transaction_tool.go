package mcp

import (
	"context"
	"fmt"
	"time"

	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func parseCommonTxFields(args map[string]any) (*transactionsv1.CreateTransactionRequest, error) {
	req := &transactionsv1.CreateTransactionRequest{}

	title, ok := args["title"].(string)
	if !ok || title == "" {
		return nil, errors.New("title is required")
	}
	req.Title = title

	dateStr, ok := args["transaction_date"].(string)
	if !ok || dateStr == "" {
		return nil, errors.New("transaction_date is required (RFC3339)")
	}
	parsed, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return nil, errors.Wrap(err, "invalid transaction_date")
	}
	req.TransactionDate = timestamppb.New(parsed)

	if notes, ok := args["notes"].(string); ok {
		req.Notes = notes
	}

	if extraRaw, ok := args["extra"].(map[string]any); ok {
		extra := make(map[string]string, len(extraRaw))
		for k, v := range extraRaw {
			s, ok := v.(string)
			if !ok {
				return nil, errors.Newf("extra[%q] must be a string", k)
			}
			extra[k] = s
		}
		req.Extra = extra
	}

	tagIDs, err := parseInt32SliceArg(args, "tag_ids")
	if err != nil {
		return nil, err
	}
	if tagIDs != nil {
		req.TagIds = tagIDs
	}

	if ref, ok := args["reference_number"].(string); ok && ref != "" {
		r := ref
		req.ReferenceNumber = &r
	}

	if irnRaw, ok := args["internal_reference_numbers"].([]any); ok {
		irn := make([]string, 0, len(irnRaw))
		for i, v := range irnRaw {
			s, ok := v.(string)
			if !ok {
				return nil, errors.Newf("internal_reference_numbers[%d] must be a string", i)
			}
			irn = append(irn, s)
		}
		req.InternalReferenceNumbers = irn
	}

	if gk, ok := args["group_key"].(string); ok && gk != "" {
		g := gk
		req.GroupKey = &g
	}

	if sr, ok := args["skip_rules"].(bool); ok {
		req.SkipRules = sr
	}

	catID, err := parseOptionalInt32PtrArg(args, "category_id")
	if err != nil {
		return nil, err
	}
	if catID != nil {
		req.CategoryId = catID
	}

	return req, nil
}

func parseInt32SliceArg(args map[string]any, key string) ([]int32, error) {
	raw, ok := args[key].([]any)
	if !ok {
		return nil, nil
	}
	out := make([]int32, 0, len(raw))
	for i, v := range raw {
		n, ok := v.(float64)
		if !ok {
			return nil, errors.Newf("%s[%d] must be a number", key, i)
		}
		out = append(out, int32(n))
	}
	return out, nil
}

func parseOptionalInt32PtrArg(args map[string]any, key string) (*int32, error) {
	v, exists := args[key]
	if !exists || v == nil {
		return nil, nil
	}
	n, ok := v.(float64)
	if !ok {
		return nil, errors.Newf("%s must be a number", key)
	}
	out := int32(n)
	return &out, nil
}

func parseExpenseFields(args map[string]any) (*transactionsv1.Expense, error) {
	srcAcc, ok := args["source_account_id"].(float64)
	if !ok {
		return nil, errors.New("source_account_id is required")
	}
	srcAmtStr, ok := args["source_amount"].(string)
	if !ok || srcAmtStr == "" {
		return nil, errors.New("source_amount is required")
	}
	if _, err := decimal.NewFromString(srcAmtStr); err != nil {
		return nil, errors.Wrap(err, "invalid source_amount")
	}
	srcCur, ok := args["source_currency"].(string)
	if !ok || srcCur == "" {
		return nil, errors.New("source_currency is required")
	}
	dstAcc, ok := args["destination_account_id"].(float64)
	if !ok {
		return nil, errors.New("destination_account_id is required")
	}
	dstAmtStr, ok := args["destination_amount"].(string)
	if !ok || dstAmtStr == "" {
		return nil, errors.New("destination_amount is required")
	}
	if _, err := decimal.NewFromString(dstAmtStr); err != nil {
		return nil, errors.Wrap(err, "invalid destination_amount")
	}
	dstCur, ok := args["destination_currency"].(string)
	if !ok || dstCur == "" {
		return nil, errors.New("destination_currency is required")
	}

	expense := &transactionsv1.Expense{
		SourceAmount:         srcAmtStr,
		SourceCurrency:       srcCur,
		SourceAccountId:      int32(srcAcc),
		DestinationAccountId: int32(dstAcc),
		DestinationAmount:    dstAmtStr,
		DestinationCurrency:  dstCur,
	}

	if fxAmt, ok := args["fx_source_amount"].(string); ok && fxAmt != "" {
		if _, err := decimal.NewFromString(fxAmt); err != nil {
			return nil, errors.Wrap(err, "invalid fx_source_amount")
		}
		v := fxAmt
		expense.FxSourceAmount = &v
	}
	if fxCur, ok := args["fx_source_currency"].(string); ok && fxCur != "" {
		v := fxCur
		expense.FxSourceCurrency = &v
	}

	return expense, nil
}

func (s *Server) handleCreateExpense(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	req, err := parseCommonTxFields(args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	expense, err := parseExpenseFields(args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	req.Transaction = &transactionsv1.CreateTransactionRequest_Expense{Expense: expense}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.TransactionSvc.Create(queryCtx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create expense: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Transaction created with id %d", resp.Transaction.Id)), nil
}

func parseIncomeFields(args map[string]any) (*transactionsv1.Income, error) {
	srcAcc, ok := args["source_account_id"].(float64)
	if !ok {
		return nil, errors.New("source_account_id is required")
	}
	srcAmtStr, ok := args["source_amount"].(string)
	if !ok || srcAmtStr == "" {
		return nil, errors.New("source_amount is required")
	}
	if _, err := decimal.NewFromString(srcAmtStr); err != nil {
		return nil, errors.Wrap(err, "invalid source_amount")
	}
	srcCur, ok := args["source_currency"].(string)
	if !ok || srcCur == "" {
		return nil, errors.New("source_currency is required")
	}
	dstAcc, ok := args["destination_account_id"].(float64)
	if !ok {
		return nil, errors.New("destination_account_id is required")
	}
	dstAmtStr, ok := args["destination_amount"].(string)
	if !ok || dstAmtStr == "" {
		return nil, errors.New("destination_amount is required")
	}
	if _, err := decimal.NewFromString(dstAmtStr); err != nil {
		return nil, errors.Wrap(err, "invalid destination_amount")
	}
	dstCur, ok := args["destination_currency"].(string)
	if !ok || dstCur == "" {
		return nil, errors.New("destination_currency is required")
	}

	return &transactionsv1.Income{
		SourceAmount:         srcAmtStr,
		SourceCurrency:       srcCur,
		SourceAccountId:      int32(srcAcc),
		DestinationAccountId: int32(dstAcc),
		DestinationAmount:    dstAmtStr,
		DestinationCurrency:  dstCur,
	}, nil
}

func (s *Server) handleCreateIncome(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	req, err := parseCommonTxFields(args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	income, err := parseIncomeFields(args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	req.Transaction = &transactionsv1.CreateTransactionRequest_Income{Income: income}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.TransactionSvc.Create(queryCtx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create income: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Transaction created with id %d", resp.Transaction.Id)), nil
}

func (s *Server) handleCreateTransfer(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError("not implemented"), nil
}

func (s *Server) handleCreateAdjustment(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError("not implemented"), nil
}

func (s *Server) handleUpdateExpense(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	idF, ok := args["id"].(float64)
	if !ok {
		return mcp.NewToolResultError("id is required"), nil
	}

	req, err := parseCommonTxFields(args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	expense, err := parseExpenseFields(args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	req.Transaction = &transactionsv1.CreateTransactionRequest_Expense{Expense: expense}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.TransactionSvc.Update(queryCtx, &transactionsv1.UpdateTransactionRequest{
		Id:          int64(idF),
		Transaction: req,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update expense: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Transaction %d updated", resp.Transaction.Id)), nil
}

func (s *Server) handleUpdateIncome(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError("not implemented"), nil
}

func (s *Server) handleUpdateTransfer(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError("not implemented"), nil
}

func (s *Server) handleUpdateAdjustment(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError("not implemented"), nil
}
