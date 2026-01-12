package mcp_test

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gomcp "github.com/ft-t/go-money/pkg/mcp"
	"github.com/ft-t/go-money/pkg/testingutils"
)

func TestNewServer_Success(t *testing.T) {
	type tc struct {
		name string
		docs string
	}

	cases := []tc{
		{
			name: "with docs",
			docs: "# Schema Documentation",
		},
		{
			name: "empty docs",
			docs: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			catSvc := NewMockCategoryService(ctrl)
			rulesSvc := NewMockRulesService(ctrl)
			dryRunSvc := NewMockDryRunService(ctrl)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        c.docs,
				CategorySvc: catSvc,
				RulesSvc:    rulesSvc,
				DryRunSvc:   dryRunSvc,
			})

			assert.NotNil(t, server)
			assert.NotNil(t, server.Handler())
			assert.NotNil(t, server.MCPServer())
		})
	}
}

func TestServer_HandleSchemaResource_Success(t *testing.T) {
	type tc struct {
		name         string
		docs         string
		expectedText string
	}

	cases := []tc{
		{
			name:         "returns docs content",
			docs:         "# Database Schema\n\nTable definitions here.",
			expectedText: "# Database Schema\n\nTable definitions here.",
		},
		{
			name:         "empty docs",
			docs:         "",
			expectedText: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			catSvc := NewMockCategoryService(ctrl)
			rulesSvc := NewMockRulesService(ctrl)
			dryRunSvc := NewMockDryRunService(ctrl)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        c.docs,
				CategorySvc: catSvc,
				RulesSvc:    rulesSvc,
				DryRunSvc:   dryRunSvc,
			})

			mcpServer := server.MCPServer()

			result := mcpServer.HandleMessage(context.Background(), []byte(`{
				"jsonrpc": "2.0",
				"id": 1,
				"method": "resources/read",
				"params": {
					"uri": "context://schema"
				}
			}`))

			require.NotNil(t, result)
		})
	}
}

func TestServer_HandleQuery_Success(t *testing.T) {
	type tc struct {
		name     string
		sql      string
		columns  []string
		rows     [][]driver.Value
		expected string
	}

	cases := []tc{
		{
			name:     "simple select",
			sql:      "SELECT id, name FROM users",
			columns:  []string{"id", "name"},
			rows:     [][]driver.Value{{1, "Alice"}, {2, "Bob"}},
			expected: "Alice",
		},
		{
			name:     "select with where",
			sql:      "SELECT balance FROM accounts WHERE id = 1",
			columns:  []string{"balance"},
			rows:     [][]driver.Value{{100.50}},
			expected: "100.5",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, mock := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			mockRows := sqlmock.NewRows(c.columns)
			for _, row := range c.rows {
				mockRows.AddRow(row...)
			}
			mock.ExpectQuery(".*").WillReturnRows(mockRows)

			catSvc := NewMockCategoryService(ctrl)
			rulesSvc := NewMockRulesService(ctrl)
			dryRunSvc := NewMockDryRunService(ctrl)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
				RulesSvc:    rulesSvc,
				DryRunSvc:   dryRunSvc,
			})

			mcpServer := server.MCPServer()

			tool := mcpServer.GetTool("query")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "query",
					Arguments: map[string]any{"sql": c.sql},
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expected)
		})
	}
}

func TestServer_HandleQuery_Failure(t *testing.T) {
	type tc struct {
		name          string
		sql           any
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing sql parameter",
			sql:           nil,
			expectedError: "sql parameter is required",
		},
		{
			name:          "empty sql",
			sql:           "",
			expectedError: "sql parameter is required",
		},
		{
			name:          "non-select query INSERT",
			sql:           "INSERT INTO users (name) VALUES ('test')",
			expectedError: "only SELECT queries are allowed",
		},
		{
			name:          "non-select query UPDATE",
			sql:           "UPDATE users SET name = 'test'",
			expectedError: "only SELECT queries are allowed",
		},
		{
			name:          "non-select query DELETE",
			sql:           "DELETE FROM users",
			expectedError: "only SELECT queries are allowed",
		},
		{
			name:          "non-select query DROP",
			sql:           "DROP TABLE users",
			expectedError: "only SELECT queries are allowed",
		},
		{
			name:          "forbidden pattern with semicolon",
			sql:           "SELECT * FROM users; DROP TABLE users",
			expectedError: "forbidden SQL operation detected",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			catSvc := NewMockCategoryService(ctrl)
			rulesSvc := NewMockRulesService(ctrl)
			dryRunSvc := NewMockDryRunService(ctrl)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
				RulesSvc:    rulesSvc,
				DryRunSvc:   dryRunSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("query")
			require.NotNil(t, tool)

			args := map[string]any{}
			if c.sql != nil {
				args["sql"] = c.sql
			}

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "query",
					Arguments: args,
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expectedError)
		})
	}
}
