package mcp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/golang/mock/gomock"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ft-t/go-money/pkg/transactions"
)

func TestServer_HandleBulkSetTransactionCategory_Success(t *testing.T) {
	int32Ptr := func(v int32) *int32 { return &v }

	type tc struct {
		name        string
		assignments []map[string]any
		expected    []transactions.CategoryAssignment
	}

	cases := []tc{
		{
			name: "set single category",
			assignments: []map[string]any{
				{"transaction_id": float64(1), "category_id": float64(5)},
			},
			expected: []transactions.CategoryAssignment{
				{TransactionID: 1, CategoryID: int32Ptr(5)},
			},
		},
		{
			name: "set multiple categories",
			assignments: []map[string]any{
				{"transaction_id": float64(1), "category_id": float64(5)},
				{"transaction_id": float64(2), "category_id": float64(10)},
			},
			expected: []transactions.CategoryAssignment{
				{TransactionID: 1, CategoryID: int32Ptr(5)},
				{TransactionID: 2, CategoryID: int32Ptr(10)},
			},
		},
		{
			name: "clear category",
			assignments: []map[string]any{
				{"transaction_id": float64(1), "category_id": nil},
			},
			expected: []transactions.CategoryAssignment{
				{TransactionID: 1, CategoryID: nil},
			},
		},
		{
			name: "mixed set and clear",
			assignments: []map[string]any{
				{"transaction_id": float64(1), "category_id": float64(5)},
				{"transaction_id": float64(2), "category_id": nil},
			},
			expected: []transactions.CategoryAssignment{
				{TransactionID: 1, CategoryID: int32Ptr(5)},
				{TransactionID: 2, CategoryID: nil},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			txSvc := NewMockTransactionService(ctrl)
			expected := c.expected
			txSvc.EXPECT().BulkSetCategory(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, got []transactions.CategoryAssignment) error {
					assert.Equal(t, expected, got)
					return nil
				})

			server := newTxServer(t, ctrl, txSvc)
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
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, fmt.Sprintf("Updated %d transactions", len(c.expected)))
		})
	}
}

func TestServer_HandleBulkSetTransactionCategory_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockTransactionService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing assignments",
			args:          map[string]any{},
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "assignments parameter is required",
		},
		{
			name:          "empty assignments",
			args:          map[string]any{"assignments": []any{}},
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "assignments parameter is required and must be a non-empty array",
		},
		{
			name:          "invalid assignment type",
			args:          map[string]any{"assignments": []any{"not an object"}},
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "assignment[0] must be an object",
		},
		{
			name: "missing transaction_id",
			args: map[string]any{"assignments": []any{
				map[string]any{"category_id": float64(5)},
			}},
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "assignment[0].transaction_id is required",
		},
		{
			name: "invalid category_id type",
			args: map[string]any{"assignments": []any{
				map[string]any{"transaction_id": float64(1), "category_id": "invalid"},
			}},
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "assignment[0].category_id must be a number or null",
		},
		{
			name: "service returns error",
			args: map[string]any{"assignments": []any{
				map[string]any{"transaction_id": float64(1), "category_id": float64(5)},
			}},
			setupMock: func(m *MockTransactionService) {
				m.EXPECT().BulkSetCategory(gomock.Any(), gomock.Any()).
					Return(errors.New("boom"))
			},
			expectedError: "failed to update transactions",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			txSvc := NewMockTransactionService(ctrl)
			c.setupMock(txSvc)

			server := newTxServer(t, ctrl, txSvc)
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
