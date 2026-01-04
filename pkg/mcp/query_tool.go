package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

const (
	queryTimeout = 30 * time.Second
	maxRows      = 1000
)

var (
	forbiddenPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^\s*(INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|TRUNCATE|GRANT|REVOKE)`),
		regexp.MustCompile(`(?i);\s*(INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|TRUNCATE|GRANT|REVOKE)`),
		regexp.MustCompile(`(?i)INTO\s+OUTFILE`),
		regexp.MustCompile(`(?i)LOAD\s+DATA`),
	}
	selectPattern = regexp.MustCompile(`(?i)^\s*SELECT\b`)
)

func (s *Server) handleQuery(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sqlQuery, ok := request.GetArguments()["sql"].(string)
	if !ok || sqlQuery == "" {
		return mcp.NewToolResultError("sql parameter is required"), nil
	}

	if err := validateQuery(sqlQuery); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var results []map[string]any

	rows, err := s.db.WithContext(queryCtx).Raw(sqlQuery).Rows()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("query error: %v", err)), nil
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("columns error: %v", err)), nil
	}

	rowCount := 0
	for rows.Next() {
		if rowCount >= maxRows {
			break
		}

		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if scanErr := rows.Scan(valuePtrs...); scanErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("scan error: %v", scanErr)), nil
		}

		row := make(map[string]any)
		for i, col := range columns {
			row[col] = convertValue(values[i])
		}

		results = append(results, row)
		rowCount++
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("rows error: %v", rowsErr)), nil
	}

	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("json error: %v", err)), nil
	}

	resultText := string(output)
	if rowCount >= maxRows {
		resultText = fmt.Sprintf("(truncated to %d rows)\n%s", maxRows, resultText)
	}

	return mcp.NewToolResultText(resultText), nil
}

func validateQuery(sql string) error {
	sql = strings.TrimSpace(sql)

	if !selectPattern.MatchString(sql) {
		return fmt.Errorf("only SELECT queries are allowed")
	}

	for _, pattern := range forbiddenPatterns {
		if pattern.MatchString(sql) {
			return fmt.Errorf("forbidden SQL operation detected")
		}
	}

	return nil
}

func convertValue(v any) any {
	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		return val.Format(time.RFC3339)
	default:
		return val
	}
}
