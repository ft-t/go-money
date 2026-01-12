package mcp_test

import (
	"context"
	"testing"

	rulesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/rules/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/golang/mock/gomock"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gomcp "github.com/ft-t/go-money/pkg/mcp"
	"github.com/ft-t/go-money/pkg/testingutils"
)

func TestServer_HandleListRules_Success(t *testing.T) {
	type tc struct {
		name     string
		rules    []*gomoneypbv1.Rule
		expected string
	}

	cases := []tc{
		{
			name: "returns rules",
			rules: []*gomoneypbv1.Rule{
				{Id: 1, Title: "Rule 1", Script: "return true"},
				{Id: 2, Title: "Rule 2", Script: "return false"},
			},
			expected: "Rule 1",
		},
		{
			name:     "no rules",
			rules:    []*gomoneypbv1.Rule{},
			expected: "No rules found",
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

			rulesSvc.EXPECT().ListRules(gomock.Any(), gomock.Any()).
				Return(&rulesv1.ListRulesResponse{Rules: c.rules}, nil)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
				RulesSvc:    rulesSvc,
				DryRunSvc:   dryRunSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("list_rules")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "list_rules",
					Arguments: map[string]any{},
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expected)
		})
	}
}

func TestServer_HandleListRules_Failure(t *testing.T) {
	type tc struct {
		name          string
		setupMock     func(*MockRulesService)
		expectedError string
	}

	cases := []tc{
		{
			name: "service error",
			setupMock: func(m *MockRulesService) {
				m.EXPECT().ListRules(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to list rules",
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

			c.setupMock(rulesSvc)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
				RulesSvc:    rulesSvc,
				DryRunSvc:   dryRunSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("list_rules")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "list_rules",
					Arguments: map[string]any{},
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expectedError)
		})
	}
}

func TestServer_HandleDryRunRule_Success(t *testing.T) {
	type tc struct {
		name        string
		txID        float64
		script      string
		title       string
		ruleApplied bool
		expected    string
	}

	cases := []tc{
		{
			name:        "rule applied",
			txID:        1,
			script:      "tx:SetTitle('New Title')\nreturn true",
			title:       "Test Rule",
			ruleApplied: true,
			expected:    `"rule_applied": true`,
		},
		{
			name:        "rule not applied",
			txID:        2,
			script:      "return false",
			title:       "Test Rule 2",
			ruleApplied: false,
			expected:    `"rule_applied": false`,
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

			dryRunSvc.EXPECT().DryRunRule(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *rulesv1.DryRunRuleRequest) (*rulesv1.DryRunRuleResponse, error) {
					assert.Equal(t, int64(c.txID), req.TransactionId)
					assert.Equal(t, c.script, req.Rule.Script)
					assert.Equal(t, c.title, req.Rule.Title)
					return &rulesv1.DryRunRuleResponse{
						RuleApplied: c.ruleApplied,
						Before:      &gomoneypbv1.Transaction{Id: int64(c.txID)},
						After:       &gomoneypbv1.Transaction{Id: int64(c.txID)},
					}, nil
				})

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
				RulesSvc:    rulesSvc,
				DryRunSvc:   dryRunSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("dry_run_rule")
			require.NotNil(t, tool)

			args := map[string]any{
				"transaction_id": c.txID,
				"script":         c.script,
				"title":          c.title,
			}

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "dry_run_rule",
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

func TestServer_HandleDryRunRule_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockDryRunService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing transaction_id",
			args:          map[string]any{"script": "return true"},
			setupMock:     func(m *MockDryRunService) {},
			expectedError: "transaction_id parameter is required",
		},
		{
			name:          "missing script",
			args:          map[string]any{"transaction_id": float64(1)},
			setupMock:     func(m *MockDryRunService) {},
			expectedError: "script parameter is required",
		},
		{
			name:          "empty script",
			args:          map[string]any{"transaction_id": float64(1), "script": ""},
			setupMock:     func(m *MockDryRunService) {},
			expectedError: "script parameter is required",
		},
		{
			name: "service error",
			args: map[string]any{"transaction_id": float64(1), "script": "return true"},
			setupMock: func(m *MockDryRunService) {
				m.EXPECT().DryRunRule(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("execution error"))
			},
			expectedError: "dry run failed",
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

			c.setupMock(dryRunSvc)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
				RulesSvc:    rulesSvc,
				DryRunSvc:   dryRunSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("dry_run_rule")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "dry_run_rule",
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

func TestServer_HandleCreateRule_Success(t *testing.T) {
	type tc struct {
		name     string
		title    string
		script   string
		respID   int32
		expected string
	}

	cases := []tc{
		{
			name:     "create rule",
			title:    "Test Rule",
			script:   "return true",
			respID:   1,
			expected: "Rule created with id 1",
		},
		{
			name:     "create rule with special chars",
			title:    "Rule: Set Category",
			script:   "tx:SetCategoryID(1)\nreturn true",
			respID:   2,
			expected: "Rule created with id 2",
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

			rulesSvc.EXPECT().CreateRule(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *rulesv1.CreateRuleRequest) (*rulesv1.CreateRuleResponse, error) {
					assert.Equal(t, c.title, req.Rule.Title)
					assert.Equal(t, c.script, req.Rule.Script)
					return &rulesv1.CreateRuleResponse{
						Rule: &gomoneypbv1.Rule{
							Id:     c.respID,
							Title:  c.title,
							Script: c.script,
						},
					}, nil
				})

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
				RulesSvc:    rulesSvc,
				DryRunSvc:   dryRunSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("create_rule")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_rule",
					Arguments: map[string]any{"title": c.title, "script": c.script},
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expected)
		})
	}
}

func TestServer_HandleCreateRule_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockRulesService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing title",
			args:          map[string]any{"script": "return true"},
			setupMock:     func(m *MockRulesService) {},
			expectedError: "title parameter is required",
		},
		{
			name:          "empty title",
			args:          map[string]any{"title": "", "script": "return true"},
			setupMock:     func(m *MockRulesService) {},
			expectedError: "title parameter is required",
		},
		{
			name:          "missing script",
			args:          map[string]any{"title": "Test"},
			setupMock:     func(m *MockRulesService) {},
			expectedError: "script parameter is required",
		},
		{
			name:          "empty script",
			args:          map[string]any{"title": "Test", "script": ""},
			setupMock:     func(m *MockRulesService) {},
			expectedError: "script parameter is required",
		},
		{
			name: "service error",
			args: map[string]any{"title": "Test", "script": "return true"},
			setupMock: func(m *MockRulesService) {
				m.EXPECT().CreateRule(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database error"))
			},
			expectedError: "failed to create rule",
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

			c.setupMock(rulesSvc)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
				RulesSvc:    rulesSvc,
				DryRunSvc:   dryRunSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("create_rule")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_rule",
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

func TestServer_HandleUpdateRule_Success(t *testing.T) {
	type tc struct {
		name     string
		ruleID   float64
		title    string
		script   string
		expected string
	}

	cases := []tc{
		{
			name:     "update rule",
			ruleID:   1,
			title:    "Updated Rule",
			script:   "return false",
			expected: "Rule 1 updated",
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

			rulesSvc.EXPECT().UpdateRule(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *rulesv1.UpdateRuleRequest) (*rulesv1.UpdateRuleResponse, error) {
					assert.Equal(t, int32(c.ruleID), req.Rule.Id)
					assert.Equal(t, c.title, req.Rule.Title)
					assert.Equal(t, c.script, req.Rule.Script)
					return &rulesv1.UpdateRuleResponse{
						Rule: &gomoneypbv1.Rule{
							Id:     int32(c.ruleID),
							Title:  c.title,
							Script: c.script,
						},
					}, nil
				})

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
				RulesSvc:    rulesSvc,
				DryRunSvc:   dryRunSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("update_rule")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "update_rule",
					Arguments: map[string]any{"id": c.ruleID, "title": c.title, "script": c.script},
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expected)
		})
	}
}

func TestServer_HandleUpdateRule_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockRulesService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing id",
			args:          map[string]any{"title": "Test", "script": "return true"},
			setupMock:     func(m *MockRulesService) {},
			expectedError: "id parameter is required",
		},
		{
			name:          "missing title",
			args:          map[string]any{"id": float64(1), "script": "return true"},
			setupMock:     func(m *MockRulesService) {},
			expectedError: "title parameter is required",
		},
		{
			name:          "missing script",
			args:          map[string]any{"id": float64(1), "title": "Test"},
			setupMock:     func(m *MockRulesService) {},
			expectedError: "script parameter is required",
		},
		{
			name: "service error",
			args: map[string]any{"id": float64(999), "title": "Test", "script": "return true"},
			setupMock: func(m *MockRulesService) {
				m.EXPECT().UpdateRule(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))
			},
			expectedError: "failed to update rule",
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

			c.setupMock(rulesSvc)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
				RulesSvc:    rulesSvc,
				DryRunSvc:   dryRunSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("update_rule")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "update_rule",
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
