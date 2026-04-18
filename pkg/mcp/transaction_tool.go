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
		req.ReferenceNumber = &ref
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
		req.GroupKey = &gk
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

type srcDstFields struct {
	SrcAcc int32
	SrcAmt string
	SrcCur string
	DstAcc int32
	DstAmt string
	DstCur string
}

func parseSrcDstFields(args map[string]any) (srcDstFields, error) {
	var f srcDstFields

	srcAcc, ok := args["source_account_id"].(float64)
	if !ok {
		return f, errors.New("source_account_id is required")
	}
	f.SrcAcc = int32(srcAcc)

	srcAmtStr, ok := args["source_amount"].(string)
	if !ok || srcAmtStr == "" {
		return f, errors.New("source_amount is required")
	}
	if _, err := decimal.NewFromString(srcAmtStr); err != nil {
		return f, errors.Wrap(err, "invalid source_amount")
	}
	f.SrcAmt = srcAmtStr

	srcCur, ok := args["source_currency"].(string)
	if !ok || srcCur == "" {
		return f, errors.New("source_currency is required")
	}
	f.SrcCur = srcCur

	dstAcc, ok := args["destination_account_id"].(float64)
	if !ok {
		return f, errors.New("destination_account_id is required")
	}
	f.DstAcc = int32(dstAcc)

	dstAmtStr, ok := args["destination_amount"].(string)
	if !ok || dstAmtStr == "" {
		return f, errors.New("destination_amount is required")
	}
	if _, err := decimal.NewFromString(dstAmtStr); err != nil {
		return f, errors.Wrap(err, "invalid destination_amount")
	}
	f.DstAmt = dstAmtStr

	dstCur, ok := args["destination_currency"].(string)
	if !ok || dstCur == "" {
		return f, errors.New("destination_currency is required")
	}
	f.DstCur = dstCur

	return f, nil
}

func parseExpenseFields(args map[string]any) (*transactionsv1.Expense, error) {
	f, err := parseSrcDstFields(args)
	if err != nil {
		return nil, err
	}

	expense := &transactionsv1.Expense{
		SourceAmount:         f.SrcAmt,
		SourceCurrency:       f.SrcCur,
		SourceAccountId:      f.SrcAcc,
		DestinationAccountId: f.DstAcc,
		DestinationAmount:    f.DstAmt,
		DestinationCurrency:  f.DstCur,
	}

	if fxAmt, ok := args["fx_source_amount"].(string); ok && fxAmt != "" {
		if _, err := decimal.NewFromString(fxAmt); err != nil {
			return nil, errors.Wrap(err, "invalid fx_source_amount")
		}
		expense.FxSourceAmount = &fxAmt
	}
	if fxCur, ok := args["fx_source_currency"].(string); ok && fxCur != "" {
		expense.FxSourceCurrency = &fxCur
	}

	return expense, nil
}

func parseIncomeFields(args map[string]any) (*transactionsv1.Income, error) {
	f, err := parseSrcDstFields(args)
	if err != nil {
		return nil, err
	}
	return &transactionsv1.Income{
		SourceAmount:         f.SrcAmt,
		SourceCurrency:       f.SrcCur,
		SourceAccountId:      f.SrcAcc,
		DestinationAccountId: f.DstAcc,
		DestinationAmount:    f.DstAmt,
		DestinationCurrency:  f.DstCur,
	}, nil
}

func parseTransferFields(args map[string]any) (*transactionsv1.TransferBetweenAccounts, error) {
	f, err := parseSrcDstFields(args)
	if err != nil {
		return nil, err
	}
	return &transactionsv1.TransferBetweenAccounts{
		SourceAmount:         f.SrcAmt,
		SourceCurrency:       f.SrcCur,
		SourceAccountId:      f.SrcAcc,
		DestinationAccountId: f.DstAcc,
		DestinationAmount:    f.DstAmt,
		DestinationCurrency:  f.DstCur,
	}, nil
}

func parseAdjustmentFields(args map[string]any) (*transactionsv1.Adjustment, error) {
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

	return &transactionsv1.Adjustment{
		DestinationAccountId: int32(dstAcc),
		DestinationAmount:    dstAmtStr,
		DestinationCurrency:  dstCur,
	}, nil
}

func (s *Server) buildCreateRequest(args map[string]any, kind string) (*transactionsv1.CreateTransactionRequest, error) {
	req, err := parseCommonTxFields(args)
	if err != nil {
		return nil, err
	}

	switch kind {
	case "expense":
		expense, err := parseExpenseFields(args)
		if err != nil {
			return nil, err
		}
		req.Transaction = &transactionsv1.CreateTransactionRequest_Expense{Expense: expense}
	case "income":
		income, err := parseIncomeFields(args)
		if err != nil {
			return nil, err
		}
		req.Transaction = &transactionsv1.CreateTransactionRequest_Income{Income: income}
	case "transfer":
		transfer, err := parseTransferFields(args)
		if err != nil {
			return nil, err
		}
		req.Transaction = &transactionsv1.CreateTransactionRequest_TransferBetweenAccounts{TransferBetweenAccounts: transfer}
	case "adjustment":
		adjustment, err := parseAdjustmentFields(args)
		if err != nil {
			return nil, err
		}
		req.Transaction = &transactionsv1.CreateTransactionRequest_Adjustment{Adjustment: adjustment}
	default:
		return nil, errors.Newf("unknown transaction kind %q", kind)
	}

	return req, nil
}

func (s *Server) handleCreate(ctx context.Context, request mcp.CallToolRequest, kind string) (*mcp.CallToolResult, error) {
	req, err := s.buildCreateRequest(request.GetArguments(), kind)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.TransactionSvc.Create(queryCtx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create %s: %v", kind, err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Transaction created with id %d", resp.Transaction.Id)), nil
}

func (s *Server) handleUpdate(ctx context.Context, request mcp.CallToolRequest, kind string) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	idF, ok := args["id"].(float64)
	if !ok {
		return mcp.NewToolResultError("id is required"), nil
	}

	req, err := s.buildCreateRequest(args, kind)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.TransactionSvc.Update(queryCtx, &transactionsv1.UpdateTransactionRequest{
		Id:          int64(idF),
		Transaction: req,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update %s: %v", kind, err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Transaction %d updated", resp.Transaction.Id)), nil
}

func (s *Server) handleCreateExpense(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreate(ctx, request, "expense")
}

func (s *Server) handleCreateIncome(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreate(ctx, request, "income")
}

func (s *Server) handleCreateTransfer(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreate(ctx, request, "transfer")
}

func (s *Server) handleCreateAdjustment(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCreate(ctx, request, "adjustment")
}

func (s *Server) handleUpdateExpense(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUpdate(ctx, request, "expense")
}

func (s *Server) handleUpdateIncome(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUpdate(ctx, request, "income")
}

func (s *Server) handleUpdateTransfer(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUpdate(ctx, request, "transfer")
}

func (s *Server) handleUpdateAdjustment(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleUpdate(ctx, request, "adjustment")
}
