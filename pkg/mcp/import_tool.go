package mcp

import (
	"context"
	"fmt"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	"github.com/cockroachdb/errors"
	"github.com/ft-t/go-money/pkg/database"
	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/protobuf/encoding/protojson"
)

func parseImportSourceArg(args map[string]any) (importv1.ImportSource, error) {
	name, _ := args["source"].(string)

	source, ok := importv1.ImportSource_value[name]
	if !ok || source == int32(importv1.ImportSource_IMPORT_SOURCE_UNSPECIFIED) {
		return 0, errors.Newf("unknown source %q", name)
	}

	return importv1.ImportSource(source), nil
}

func (s *Server) handleParseImport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	source, err := parseImportSourceArg(args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	content, err := parseStringSliceArg(args, "content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if len(content) == 0 {
		return mcp.NewToolResultError("content is required and must be a non-empty array"), nil
	}

	skipRules, _ := args["skip_rules"].(bool)
	treatDatesAsUtc, _ := args["treat_dates_as_utc"].(bool)

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.ImportSvc.Parse(queryCtx, &importv1.ParseTransactionsRequest{
		Content:         content,
		Source:          source,
		SkipRules:       skipRules,
		TreatDatesAsUtc: treatDatesAsUtc,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to parse import: %v", err)), nil
	}

	result, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(resp)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to format parsed import: %v", err)), nil
	}

	return mcp.NewToolResultText(string(result)), nil
}

func (s *Server) handleMarkImportIgnored(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	source, err := parseImportSourceArg(args)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	refs, err := parseStringSliceArg(args, "reference_numbers")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	req := &importv1.MarkTransactionsIgnoredRequest{
		ImportSource:     source,
		ReferenceNumbers: refs,
	}
	if reason, ok := args["reason"].(string); ok && reason != "" {
		req.Reason = &reason
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	queryCtx = database.WithContext(queryCtx, s.db)

	resp, err := s.cfg.ImportSvc.MarkTransactionsIgnored(queryCtx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to mark import rows ignored: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Ignored %d new reference numbers", resp.IgnoredCount)), nil
}
