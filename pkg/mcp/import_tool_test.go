package mcp_test

import (
	"context"
	"testing"

	importv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/import/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/golang/mock/gomock"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gomcp "github.com/ft-t/go-money/pkg/mcp"
	"github.com/ft-t/go-money/pkg/testingutils"
)

func newImportServer(t *testing.T, ctrl *gomock.Controller, importSvc *MockImportService) *gomcp.Server {
	t.Helper()
	gormDB, mockDB, _ := testingutils.GormMock()
	t.Cleanup(func() { _ = mockDB.Close() })

	return gomcp.NewServer(&gomcp.ServerConfig{
		DB:             gormDB,
		Docs:           "test docs",
		CategorySvc:    NewMockCategoryService(ctrl),
		RulesSvc:       NewMockRulesService(ctrl),
		DryRunSvc:      NewMockDryRunService(ctrl),
		TagsSvc:        NewMockTagsService(ctrl),
		TransactionSvc: NewMockTransactionService(ctrl),
		CurrencySvc:    NewMockCurrencyConverterService(ctrl),
		ImportSvc:      importSvc,
	})
}

func callTool(t *testing.T, server *gomcp.Server, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()

	tool := server.MCPServer().GetTool(name)
	require.NotNil(t, tool)

	result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: name, Arguments: args},
	})
	require.NoError(t, err)

	return result
}

func TestServer_HandleParseImport_Success(t *testing.T) {
	type tc struct {
		name     string
		args     map[string]any
		verify   func(t *testing.T, req *importv1.ParseTransactionsRequest)
		expected string
	}

	cases := []tc{
		{
			name: "passes files and source, returns reference numbers",
			args: map[string]any{
				"source":  "IMPORT_SOURCE_BNP_PARIBAS_POLSKA",
				"content": []any{"ZmlsZTE=", "ZmlsZTI="},
			},
			verify: func(t *testing.T, req *importv1.ParseTransactionsRequest) {
				assert.Equal(t, importv1.ImportSource_IMPORT_SOURCE_BNP_PARIBAS_POLSKA, req.Source)
				assert.Equal(t, []string{"ZmlsZTE=", "ZmlsZTI="}, req.Content)
			},
			expected: `"internal_reference_numbers":["IMPORT_SOURCE_BNP_PARIBAS_POLSKA_abc"]`,
		},
		{
			name: "forwards skip_rules and treat_dates_as_utc",
			args: map[string]any{
				"source":             "IMPORT_SOURCE_REVOLUT",
				"content":            []any{"raw csv"},
				"skip_rules":         true,
				"treat_dates_as_utc": true,
			},
			verify: func(t *testing.T, req *importv1.ParseTransactionsRequest) {
				assert.True(t, req.SkipRules && req.TreatDatesAsUtc)
			},
			expected: `"duplicate_transaction_id":"42"`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			importSvc := NewMockImportService(ctrl)
			importSvc.EXPECT().Parse(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *importv1.ParseTransactionsRequest) (*importv1.ParseTransactionsResponse, error) {
					c.verify(t, req)
					dupID := int64(42)
					return &importv1.ParseTransactionsResponse{
						Transactions: []*importv1.ParseTransactionsResponse_ParsedTransaction{
							{
								Transaction: &gomoneypbv1.Transaction{
									Title:                    "Cash deposit",
									InternalReferenceNumbers: []string{"IMPORT_SOURCE_BNP_PARIBAS_POLSKA_abc"},
								},
								DuplicateTransactionId: &dupID,
							},
						},
					}, nil
				})

			result := callTool(t, newImportServer(t, ctrl, importSvc), "parse_import", c.args)

			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expected)
		})
	}
}

func TestServer_HandleParseImport_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockImportService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "unknown source",
			args:          map[string]any{"source": "BNP", "content": []any{"x"}},
			setupMock:     func(m *MockImportService) {},
			expectedError: `unknown source "BNP"`,
		},
		{
			name:          "unspecified source",
			args:          map[string]any{"source": "IMPORT_SOURCE_UNSPECIFIED", "content": []any{"x"}},
			setupMock:     func(m *MockImportService) {},
			expectedError: "unknown source",
		},
		{
			name:          "missing content",
			args:          map[string]any{"source": "IMPORT_SOURCE_MBANK"},
			setupMock:     func(m *MockImportService) {},
			expectedError: "content is required",
		},
		{
			name: "service error",
			args: map[string]any{"source": "IMPORT_SOURCE_MBANK", "content": []any{"x"}},
			setupMock: func(m *MockImportService) {
				m.EXPECT().Parse(gomock.Any(), gomock.Any()).Return(nil, errors.New("bad file"))
			},
			expectedError: "failed to parse import: bad file",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			importSvc := NewMockImportService(ctrl)
			c.setupMock(importSvc)

			result := callTool(t, newImportServer(t, ctrl, importSvc), "parse_import", c.args)

			assert.True(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expectedError)
		})
	}
}

func TestServer_HandleMarkImportIgnored_Success(t *testing.T) {
	type tc struct {
		name   string
		args   map[string]any
		verify func(t *testing.T, req *importv1.MarkTransactionsIgnoredRequest)
	}

	cases := []tc{
		{
			name: "single row",
			args: map[string]any{
				"source":            "IMPORT_SOURCE_BNP_PARIBAS_POLSKA",
				"reference_numbers": []any{"IMPORT_SOURCE_BNP_PARIBAS_POLSKA_abc"},
			},
			verify: func(t *testing.T, req *importv1.MarkTransactionsIgnoredRequest) {
				assert.Equal(t, []string{"IMPORT_SOURCE_BNP_PARIBAS_POLSKA_abc"}, req.ReferenceNumbers)
			},
		},
		{
			name: "with reason",
			args: map[string]any{
				"source":            "IMPORT_SOURCE_BNP_PARIBAS_POLSKA",
				"reference_numbers": []any{"a"},
				"reason":            "created by agent",
			},
			verify: func(t *testing.T, req *importv1.MarkTransactionsIgnoredRequest) {
				assert.Equal(t, "created by agent", req.GetReason())
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			importSvc := NewMockImportService(ctrl)
			importSvc.EXPECT().MarkTransactionsIgnored(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *importv1.MarkTransactionsIgnoredRequest) (*importv1.MarkTransactionsIgnoredResponse, error) {
					c.verify(t, req)
					return &importv1.MarkTransactionsIgnoredResponse{IgnoredCount: 1}, nil
				})

			result := callTool(t, newImportServer(t, ctrl, importSvc), "mark_import_ignored", c.args)

			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "Ignored 1 new reference numbers")
		})
	}
}

func TestServer_HandleMarkImportIgnored_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockImportService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "unknown source",
			args:          map[string]any{"source": "", "reference_numbers": []any{"a"}},
			setupMock:     func(m *MockImportService) {},
			expectedError: "unknown source",
		},
		{
			name:          "non-string reference number",
			args:          map[string]any{"source": "IMPORT_SOURCE_MBANK", "reference_numbers": []any{float64(1)}},
			setupMock:     func(m *MockImportService) {},
			expectedError: "reference_numbers[0] must be a string",
		},
		{
			name: "service error",
			args: map[string]any{"source": "IMPORT_SOURCE_MBANK", "reference_numbers": []any{" "}},
			setupMock: func(m *MockImportService) {
				m.EXPECT().MarkTransactionsIgnored(gomock.Any(), gomock.Any()).Return(nil, errors.New("no valid reference numbers to ignore"))
			},
			expectedError: "failed to mark import rows ignored",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			importSvc := NewMockImportService(ctrl)
			c.setupMock(importSvc)

			result := callTool(t, newImportServer(t, ctrl, importSvc), "mark_import_ignored", c.args)

			assert.True(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expectedError)
		})
	}
}
