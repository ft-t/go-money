package mcp_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gomcp "github.com/ft-t/go-money/pkg/mcp"
	"github.com/ft-t/go-money/pkg/testingutils"
)

func TestServer_HandleSetTransactionCategory_Success(t *testing.T) {
	type tc struct {
		name       string
		txID       float64
		categoryID any
		expected   string
	}

	cases := []tc{
		{
			name:       "set category",
			txID:       1,
			categoryID: float64(5),
			expected:   "Transaction 1 category set to 5",
		},
		{
			name:       "clear category with nil",
			txID:       2,
			categoryID: nil,
			expected:   "Transaction 2 category cleared",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, mock := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			selectRows := sqlmock.NewRows([]string{
				"id", "source_amount", "source_currency", "destination_amount",
				"destination_currency", "source_account_id", "destination_account_id",
				"category_id", "transaction_type", "created_at", "updated_at",
			}).AddRow(
				int64(c.txID), nil, "USD", nil, "USD", 1, 2, nil, 3,
				time.Now(), time.Now(),
			)
			mock.ExpectQuery("SELECT \\* FROM \"transactions\"").WillReturnRows(selectRows)
			mock.ExpectBegin()
			mock.ExpectExec("UPDATE \"transactions\"").WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			catSvc := NewMockCategoryService(ctrl)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("set_transaction_category")
			require.NotNil(t, tool)

			args := map[string]any{"transaction_id": c.txID}
			if c.categoryID != nil {
				args["category_id"] = c.categoryID
			}

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "set_transaction_category",
					Arguments: args,
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expected)
		})
	}
}

func TestServer_HandleSetTransactionCategory_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(sqlmock.Sqlmock)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing transaction_id",
			args:          map[string]any{},
			setupMock:     func(m sqlmock.Sqlmock) {},
			expectedError: "transaction_id parameter is required",
		},
		{
			name:          "invalid category_id type",
			args:          map[string]any{"transaction_id": float64(1), "category_id": "invalid"},
			setupMock:     func(m sqlmock.Sqlmock) {},
			expectedError: "category_id must be a number or null",
		},
		{
			name: "transaction not found",
			args: map[string]any{"transaction_id": float64(999)},
			setupMock: func(m sqlmock.Sqlmock) {
				m.ExpectQuery("SELECT \\* FROM \"transactions\"").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			},
			expectedError: "transaction not found",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, mock := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			c.setupMock(mock)

			catSvc := NewMockCategoryService(ctrl)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("set_transaction_category")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "set_transaction_category",
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
