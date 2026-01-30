package mcp_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gomcp "github.com/ft-t/go-money/pkg/mcp"
	"github.com/ft-t/go-money/pkg/testingutils"
)

func TestServer_HandleBulkSetTransactionCategory_Success(t *testing.T) {
	type tc struct {
		name        string
		assignments []map[string]any
	}

	cases := []tc{
		{
			name: "set single category",
			assignments: []map[string]any{
				{"transaction_id": float64(1), "category_id": float64(5)},
			},
		},
		{
			name: "set multiple categories",
			assignments: []map[string]any{
				{"transaction_id": float64(1), "category_id": float64(5)},
				{"transaction_id": float64(2), "category_id": float64(10)},
			},
		},
		{
			name: "clear category",
			assignments: []map[string]any{
				{"transaction_id": float64(1), "category_id": nil},
			},
		},
		{
			name: "mixed set and clear",
			assignments: []map[string]any{
				{"transaction_id": float64(1), "category_id": float64(5)},
				{"transaction_id": float64(2), "category_id": nil},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, mock := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			for range c.assignments {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE \"transactions\"").WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: NewMockCategoryService(ctrl),
				RulesSvc:    NewMockRulesService(ctrl),
				DryRunSvc:   NewMockDryRunService(ctrl),
			})

			tool := server.MCPServer().GetTool("bulk_set_transaction_category")
			require.NotNil(t, tool)

			assignmentsAny := make([]any, len(c.assignments))
			for i, a := range c.assignments {
				assignmentsAny[i] = a
			}

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "bulk_set_transaction_category",
					Arguments: map[string]any{"assignments": assignmentsAny},
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "Updated")
		})
	}
}

func TestServer_HandleBulkSetTransactionCategory_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing assignments",
			args:          map[string]any{},
			expectedError: "assignments parameter is required",
		},
		{
			name:          "empty assignments",
			args:          map[string]any{"assignments": []any{}},
			expectedError: "assignments parameter is required and must be a non-empty array",
		},
		{
			name:          "invalid assignment type",
			args:          map[string]any{"assignments": []any{"not an object"}},
			expectedError: "assignment[0] must be an object",
		},
		{
			name: "missing transaction_id",
			args: map[string]any{"assignments": []any{
				map[string]any{"category_id": float64(5)},
			}},
			expectedError: "assignment[0].transaction_id is required",
		},
		{
			name: "invalid category_id type",
			args: map[string]any{"assignments": []any{
				map[string]any{"transaction_id": float64(1), "category_id": "invalid"},
			}},
			expectedError: "assignment[0].category_id must be a number or null",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: NewMockCategoryService(ctrl),
				RulesSvc:    NewMockRulesService(ctrl),
				DryRunSvc:   NewMockDryRunService(ctrl),
			})

			tool := server.MCPServer().GetTool("bulk_set_transaction_category")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "bulk_set_transaction_category",
					Arguments: c.args,
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expectedError)
		})
	}
}
