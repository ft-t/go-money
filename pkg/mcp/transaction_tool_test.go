package mcp_test

import (
	"context"
	"testing"

	transactionsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/golang/mock/gomock"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gomcp "github.com/ft-t/go-money/pkg/mcp"
	"github.com/ft-t/go-money/pkg/testingutils"
)

func newTxServer(t *testing.T, ctrl *gomock.Controller, txSvc *MockTransactionService) *gomcp.Server {
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
		TransactionSvc: txSvc,
		CurrencySvc:    NewMockCurrencyConverterService(ctrl),
	})
}

func TestServer_HandleCreateExpense_Success(t *testing.T) {
	type tc struct {
		name   string
		args   map[string]any
		respID int64
		verify func(t *testing.T, req *transactionsv1.CreateTransactionRequest)
	}

	cases := []tc{
		{
			name: "minimal valid expense",
			args: map[string]any{
				"transaction_date":       "2024-01-15T10:00:00Z",
				"title":                  "Lunch",
				"source_account_id":      float64(1),
				"source_amount":          "-15.50",
				"source_currency":        "USD",
				"destination_account_id": float64(2),
				"destination_amount":     "15.50",
				"destination_currency":   "USD",
			},
			respID: 42,
			verify: func(t *testing.T, req *transactionsv1.CreateTransactionRequest) {
				assert.Equal(t, "Lunch", req.Title)
				assert.Equal(t, int64(1705312800), req.TransactionDate.Seconds)
				exp := req.GetExpense()
				require.NotNil(t, exp)
				assert.Equal(t, "-15.50", exp.SourceAmount)
				assert.Equal(t, "USD", exp.SourceCurrency)
				assert.Equal(t, int32(1), exp.SourceAccountId)
				assert.Equal(t, int32(2), exp.DestinationAccountId)
				assert.Equal(t, "15.50", exp.DestinationAmount)
				assert.Equal(t, "USD", exp.DestinationCurrency)
				assert.Nil(t, exp.FxSourceAmount)
				assert.Nil(t, exp.FxSourceCurrency)
				assert.Nil(t, req.CategoryId)
				assert.Empty(t, req.TagIds)
				assert.Empty(t, req.Notes)
				assert.Empty(t, req.Extra)
				assert.Nil(t, req.ReferenceNumber)
				assert.Empty(t, req.InternalReferenceNumbers)
				assert.Nil(t, req.GroupKey)
				assert.False(t, req.SkipRules)
			},
		},
		{
			name: "with fx source amount and currency",
			args: map[string]any{
				"transaction_date":       "2024-02-01T00:00:00Z",
				"title":                  "Hotel",
				"source_account_id":      float64(3),
				"source_amount":          "-100.00",
				"source_currency":        "USD",
				"destination_account_id": float64(4),
				"destination_amount":     "92.00",
				"destination_currency":   "EUR",
				"fx_source_amount":       "-100.00",
				"fx_source_currency":     "USD",
			},
			respID: 7,
			verify: func(t *testing.T, req *transactionsv1.CreateTransactionRequest) {
				exp := req.GetExpense()
				require.NotNil(t, exp)
				require.NotNil(t, exp.FxSourceAmount)
				require.NotNil(t, exp.FxSourceCurrency)
				assert.Equal(t, "-100.00", *exp.FxSourceAmount)
				assert.Equal(t, "USD", *exp.FxSourceCurrency)
				assert.Equal(t, "EUR", exp.DestinationCurrency)
			},
		},
		{
			name: "full optional fields",
			args: map[string]any{
				"transaction_date":           "2024-03-10T12:30:00Z",
				"title":                      "Groceries",
				"source_account_id":          float64(10),
				"source_amount":              "-55.25",
				"source_currency":            "USD",
				"destination_account_id":     float64(20),
				"destination_amount":         "55.25",
				"destination_currency":       "USD",
				"notes":                      "weekly shopping",
				"extra":                      map[string]any{"store": "Trader Joes", "tx": "abc"},
				"tag_ids":                    []any{float64(1), float64(2), float64(3)},
				"reference_number":           "REF-123",
				"internal_reference_numbers": []any{"INT-1", "INT-2"},
				"group_key":                  "grp-xyz",
				"skip_rules":                 true,
				"category_id":                float64(99),
			},
			respID: 100,
			verify: func(t *testing.T, req *transactionsv1.CreateTransactionRequest) {
				assert.Equal(t, "Groceries", req.Title)
				assert.Equal(t, "weekly shopping", req.Notes)
				assert.Equal(t, map[string]string{"store": "Trader Joes", "tx": "abc"}, req.Extra)
				assert.Equal(t, []int32{1, 2, 3}, req.TagIds)
				require.NotNil(t, req.ReferenceNumber)
				assert.Equal(t, "REF-123", *req.ReferenceNumber)
				assert.Equal(t, []string{"INT-1", "INT-2"}, req.InternalReferenceNumbers)
				require.NotNil(t, req.GroupKey)
				assert.Equal(t, "grp-xyz", *req.GroupKey)
				assert.True(t, req.SkipRules)
				require.NotNil(t, req.CategoryId)
				assert.Equal(t, int32(99), *req.CategoryId)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			txSvc := NewMockTransactionService(ctrl)
			txSvc.EXPECT().Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *transactionsv1.CreateTransactionRequest) (*transactionsv1.CreateTransactionResponse, error) {
					c.verify(t, req)
					return &transactionsv1.CreateTransactionResponse{
						Transaction: &gomoneypbv1.Transaction{Id: c.respID},
					}, nil
				})

			server := newTxServer(t, ctrl, txSvc)
			tool := server.MCPServer().GetTool("create_expense")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_expense",
					Arguments: c.args,
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "Transaction created with id")
		})
	}
}

func TestServer_HandleCreateExpense_Failure(t *testing.T) {
	baseArgs := func() map[string]any {
		return map[string]any{
			"transaction_date":       "2024-01-15T10:00:00Z",
			"title":                  "Lunch",
			"source_account_id":      float64(1),
			"source_amount":          "-15.50",
			"source_currency":        "USD",
			"destination_account_id": float64(2),
			"destination_amount":     "15.50",
			"destination_currency":   "USD",
		}
	}

	argsWith := func(overrides map[string]any, remove ...string) map[string]any {
		args := baseArgs()
		for k, v := range overrides {
			args[k] = v
		}
		for _, k := range remove {
			delete(args, k)
		}
		return args
	}

	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockTransactionService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing transaction_date",
			args:          argsWith(nil, "transaction_date"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "transaction_date is required",
		},
		{
			name:          "missing title",
			args:          argsWith(nil, "title"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "title is required",
		},
		{
			name:          "invalid transaction_date",
			args:          argsWith(map[string]any{"transaction_date": "not a date"}),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "invalid transaction_date",
		},
		{
			name:          "invalid source_amount",
			args:          argsWith(map[string]any{"source_amount": "abc"}),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "invalid source_amount",
		},
		{
			name:          "invalid fx_source_amount",
			args:          argsWith(map[string]any{"fx_source_amount": "abc", "fx_source_currency": "USD"}),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "invalid fx_source_amount",
		},
		{
			name: "service returns error",
			args: baseArgs(),
			setupMock: func(m *MockTransactionService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db down"))
			},
			expectedError: "failed to create expense",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			txSvc := NewMockTransactionService(ctrl)
			c.setupMock(txSvc)

			server := newTxServer(t, ctrl, txSvc)
			tool := server.MCPServer().GetTool("create_expense")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_expense",
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

func TestServer_HandleCreateIncome_Success(t *testing.T) {
	type tc struct {
		name   string
		args   map[string]any
		respID int64
		verify func(t *testing.T, req *transactionsv1.CreateTransactionRequest)
	}

	cases := []tc{
		{
			name: "minimal valid income",
			args: map[string]any{
				"transaction_date":       "2024-01-15T10:00:00Z",
				"title":                  "Salary",
				"source_account_id":      float64(1),
				"source_amount":          "-1000.00",
				"source_currency":        "USD",
				"destination_account_id": float64(2),
				"destination_amount":     "1000.00",
				"destination_currency":   "USD",
			},
			respID: 11,
			verify: func(t *testing.T, req *transactionsv1.CreateTransactionRequest) {
				assert.Equal(t, "Salary", req.Title)
				assert.Equal(t, int64(1705312800), req.TransactionDate.Seconds)
				income := req.GetIncome()
				require.NotNil(t, income)
				assert.Equal(t, "-1000.00", income.SourceAmount)
				assert.Equal(t, "USD", income.SourceCurrency)
				assert.Equal(t, int32(1), income.SourceAccountId)
				assert.Equal(t, int32(2), income.DestinationAccountId)
				assert.Equal(t, "1000.00", income.DestinationAmount)
				assert.Equal(t, "USD", income.DestinationCurrency)
				assert.Nil(t, req.CategoryId)
				assert.Empty(t, req.TagIds)
				assert.Empty(t, req.Notes)
				assert.Empty(t, req.Extra)
				assert.Nil(t, req.ReferenceNumber)
				assert.Empty(t, req.InternalReferenceNumbers)
				assert.Nil(t, req.GroupKey)
				assert.False(t, req.SkipRules)
			},
		},
		{
			name: "full optional fields",
			args: map[string]any{
				"transaction_date":           "2024-03-10T12:30:00Z",
				"title":                      "Bonus",
				"source_account_id":          float64(10),
				"source_amount":              "-500.25",
				"source_currency":            "USD",
				"destination_account_id":     float64(20),
				"destination_amount":         "500.25",
				"destination_currency":       "USD",
				"notes":                      "q1 bonus",
				"extra":                      map[string]any{"source": "payroll"},
				"tag_ids":                    []any{float64(7), float64(8)},
				"reference_number":           "REF-INC",
				"internal_reference_numbers": []any{"INT-I1"},
				"group_key":                  "grp-inc",
				"skip_rules":                 true,
				"category_id":                float64(42),
			},
			respID: 12,
			verify: func(t *testing.T, req *transactionsv1.CreateTransactionRequest) {
				income := req.GetIncome()
				require.NotNil(t, income)
				assert.Equal(t, int32(10), income.SourceAccountId)
				assert.Equal(t, int32(20), income.DestinationAccountId)
				assert.Equal(t, "-500.25", income.SourceAmount)
				assert.Equal(t, "500.25", income.DestinationAmount)
				assert.Equal(t, "q1 bonus", req.Notes)
				assert.Equal(t, map[string]string{"source": "payroll"}, req.Extra)
				assert.Equal(t, []int32{7, 8}, req.TagIds)
				require.NotNil(t, req.ReferenceNumber)
				assert.Equal(t, "REF-INC", *req.ReferenceNumber)
				assert.Equal(t, []string{"INT-I1"}, req.InternalReferenceNumbers)
				require.NotNil(t, req.GroupKey)
				assert.Equal(t, "grp-inc", *req.GroupKey)
				assert.True(t, req.SkipRules)
				require.NotNil(t, req.CategoryId)
				assert.Equal(t, int32(42), *req.CategoryId)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			txSvc := NewMockTransactionService(ctrl)
			txSvc.EXPECT().Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *transactionsv1.CreateTransactionRequest) (*transactionsv1.CreateTransactionResponse, error) {
					c.verify(t, req)
					return &transactionsv1.CreateTransactionResponse{
						Transaction: &gomoneypbv1.Transaction{Id: c.respID},
					}, nil
				})

			server := newTxServer(t, ctrl, txSvc)
			tool := server.MCPServer().GetTool("create_income")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_income",
					Arguments: c.args,
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "Transaction created with id")
		})
	}
}

func TestServer_HandleCreateIncome_Failure(t *testing.T) {
	baseArgs := func() map[string]any {
		return map[string]any{
			"transaction_date":       "2024-01-15T10:00:00Z",
			"title":                  "Salary",
			"source_account_id":      float64(1),
			"source_amount":          "-1000.00",
			"source_currency":        "USD",
			"destination_account_id": float64(2),
			"destination_amount":     "1000.00",
			"destination_currency":   "USD",
		}
	}

	argsWith := func(overrides map[string]any, remove ...string) map[string]any {
		args := baseArgs()
		for k, v := range overrides {
			args[k] = v
		}
		for _, k := range remove {
			delete(args, k)
		}
		return args
	}

	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockTransactionService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing transaction_date",
			args:          argsWith(nil, "transaction_date"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "transaction_date is required",
		},
		{
			name:          "missing title",
			args:          argsWith(nil, "title"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "title is required",
		},
		{
			name:          "missing source_account_id",
			args:          argsWith(nil, "source_account_id"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "source_account_id is required",
		},
		{
			name:          "missing destination_amount",
			args:          argsWith(nil, "destination_amount"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "destination_amount is required",
		},
		{
			name:          "invalid source_amount",
			args:          argsWith(map[string]any{"source_amount": "not-a-decimal"}),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "invalid source_amount",
		},
		{
			name:          "invalid destination_amount",
			args:          argsWith(map[string]any{"destination_amount": "not-a-decimal"}),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "invalid destination_amount",
		},
		{
			name: "service returns error",
			args: baseArgs(),
			setupMock: func(m *MockTransactionService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db down"))
			},
			expectedError: "failed to create income",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			txSvc := NewMockTransactionService(ctrl)
			c.setupMock(txSvc)

			server := newTxServer(t, ctrl, txSvc)
			tool := server.MCPServer().GetTool("create_income")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_income",
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

func TestServer_HandleCreateTransfer_Success(t *testing.T) {
	type tc struct {
		name   string
		args   map[string]any
		respID int64
		verify func(t *testing.T, req *transactionsv1.CreateTransactionRequest)
	}

	cases := []tc{
		{
			name: "minimal valid transfer",
			args: map[string]any{
				"transaction_date":       "2024-01-15T10:00:00Z",
				"title":                  "Move cash",
				"source_account_id":      float64(3),
				"source_amount":          "-250.00",
				"source_currency":        "USD",
				"destination_account_id": float64(4),
				"destination_amount":     "250.00",
				"destination_currency":   "USD",
			},
			respID: 21,
			verify: func(t *testing.T, req *transactionsv1.CreateTransactionRequest) {
				assert.Equal(t, "Move cash", req.Title)
				assert.Equal(t, int64(1705312800), req.TransactionDate.Seconds)
				tr := req.GetTransferBetweenAccounts()
				require.NotNil(t, tr)
				assert.Equal(t, "-250.00", tr.SourceAmount)
				assert.Equal(t, "USD", tr.SourceCurrency)
				assert.Equal(t, int32(3), tr.SourceAccountId)
				assert.Equal(t, int32(4), tr.DestinationAccountId)
				assert.Equal(t, "250.00", tr.DestinationAmount)
				assert.Equal(t, "USD", tr.DestinationCurrency)
				assert.Nil(t, req.CategoryId)
				assert.Empty(t, req.TagIds)
				assert.Empty(t, req.Notes)
				assert.Empty(t, req.Extra)
				assert.Nil(t, req.ReferenceNumber)
				assert.Empty(t, req.InternalReferenceNumbers)
				assert.Nil(t, req.GroupKey)
				assert.False(t, req.SkipRules)
			},
		},
		{
			name: "full optional fields",
			args: map[string]any{
				"transaction_date":           "2024-03-10T12:30:00Z",
				"title":                      "Savings move",
				"source_account_id":          float64(30),
				"source_amount":              "-1200.50",
				"source_currency":            "USD",
				"destination_account_id":     float64(31),
				"destination_amount":         "1100.00",
				"destination_currency":       "EUR",
				"notes":                      "cross-currency move",
				"extra":                      map[string]any{"batch": "b-1"},
				"tag_ids":                    []any{float64(11)},
				"reference_number":           "REF-TR",
				"internal_reference_numbers": []any{"INT-T1", "INT-T2"},
				"group_key":                  "grp-tr",
				"skip_rules":                 true,
				"category_id":                float64(33),
			},
			respID: 22,
			verify: func(t *testing.T, req *transactionsv1.CreateTransactionRequest) {
				tr := req.GetTransferBetweenAccounts()
				require.NotNil(t, tr)
				assert.Equal(t, int32(30), tr.SourceAccountId)
				assert.Equal(t, int32(31), tr.DestinationAccountId)
				assert.Equal(t, "-1200.50", tr.SourceAmount)
				assert.Equal(t, "1100.00", tr.DestinationAmount)
				assert.Equal(t, "USD", tr.SourceCurrency)
				assert.Equal(t, "EUR", tr.DestinationCurrency)
				assert.Equal(t, "cross-currency move", req.Notes)
				assert.Equal(t, map[string]string{"batch": "b-1"}, req.Extra)
				assert.Equal(t, []int32{11}, req.TagIds)
				require.NotNil(t, req.ReferenceNumber)
				assert.Equal(t, "REF-TR", *req.ReferenceNumber)
				assert.Equal(t, []string{"INT-T1", "INT-T2"}, req.InternalReferenceNumbers)
				require.NotNil(t, req.GroupKey)
				assert.Equal(t, "grp-tr", *req.GroupKey)
				assert.True(t, req.SkipRules)
				require.NotNil(t, req.CategoryId)
				assert.Equal(t, int32(33), *req.CategoryId)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			txSvc := NewMockTransactionService(ctrl)
			txSvc.EXPECT().Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *transactionsv1.CreateTransactionRequest) (*transactionsv1.CreateTransactionResponse, error) {
					c.verify(t, req)
					return &transactionsv1.CreateTransactionResponse{
						Transaction: &gomoneypbv1.Transaction{Id: c.respID},
					}, nil
				})

			server := newTxServer(t, ctrl, txSvc)
			tool := server.MCPServer().GetTool("create_transfer")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_transfer",
					Arguments: c.args,
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "Transaction created with id")
		})
	}
}

func TestServer_HandleCreateTransfer_Failure(t *testing.T) {
	baseArgs := func() map[string]any {
		return map[string]any{
			"transaction_date":       "2024-01-15T10:00:00Z",
			"title":                  "Move cash",
			"source_account_id":      float64(3),
			"source_amount":          "-250.00",
			"source_currency":        "USD",
			"destination_account_id": float64(4),
			"destination_amount":     "250.00",
			"destination_currency":   "USD",
		}
	}

	argsWith := func(overrides map[string]any, remove ...string) map[string]any {
		args := baseArgs()
		for k, v := range overrides {
			args[k] = v
		}
		for _, k := range remove {
			delete(args, k)
		}
		return args
	}

	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockTransactionService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing transaction_date",
			args:          argsWith(nil, "transaction_date"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "transaction_date is required",
		},
		{
			name:          "missing title",
			args:          argsWith(nil, "title"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "title is required",
		},
		{
			name:          "missing source_account_id",
			args:          argsWith(nil, "source_account_id"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "source_account_id is required",
		},
		{
			name:          "missing destination_amount",
			args:          argsWith(nil, "destination_amount"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "destination_amount is required",
		},
		{
			name:          "invalid source_amount",
			args:          argsWith(map[string]any{"source_amount": "not-a-decimal"}),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "invalid source_amount",
		},
		{
			name:          "invalid destination_amount",
			args:          argsWith(map[string]any{"destination_amount": "not-a-decimal"}),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "invalid destination_amount",
		},
		{
			name: "service returns error",
			args: baseArgs(),
			setupMock: func(m *MockTransactionService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db down"))
			},
			expectedError: "failed to create transfer",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			txSvc := NewMockTransactionService(ctrl)
			c.setupMock(txSvc)

			server := newTxServer(t, ctrl, txSvc)
			tool := server.MCPServer().GetTool("create_transfer")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_transfer",
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

func TestServer_HandleCreateAdjustment_Success(t *testing.T) {
	type tc struct {
		name   string
		args   map[string]any
		respID int64
		verify func(t *testing.T, req *transactionsv1.CreateTransactionRequest)
	}

	cases := []tc{
		{
			name: "minimal valid adjustment",
			args: map[string]any{
				"transaction_date":       "2024-01-15T10:00:00Z",
				"title":                  "Balance fix",
				"destination_account_id": float64(5),
				"destination_amount":     "12.34",
				"destination_currency":   "USD",
			},
			respID: 31,
			verify: func(t *testing.T, req *transactionsv1.CreateTransactionRequest) {
				assert.Equal(t, "Balance fix", req.Title)
				assert.Equal(t, int64(1705312800), req.TransactionDate.Seconds)
				adj := req.GetAdjustment()
				require.NotNil(t, adj)
				assert.Equal(t, int32(5), adj.DestinationAccountId)
				assert.Equal(t, "12.34", adj.DestinationAmount)
				assert.Equal(t, "USD", adj.DestinationCurrency)
				assert.Nil(t, req.CategoryId)
				assert.Empty(t, req.TagIds)
				assert.Empty(t, req.Notes)
			},
		},
		{
			name: "full optional fields",
			args: map[string]any{
				"transaction_date":       "2024-03-10T12:30:00Z",
				"title":                  "Audit adj",
				"destination_account_id": float64(50),
				"destination_amount":     "-9.99",
				"destination_currency":   "EUR",
				"notes":                  "audit correction",
				"tag_ids":                []any{float64(99)},
				"category_id":            float64(8),
			},
			respID: 32,
			verify: func(t *testing.T, req *transactionsv1.CreateTransactionRequest) {
				adj := req.GetAdjustment()
				require.NotNil(t, adj)
				assert.Equal(t, int32(50), adj.DestinationAccountId)
				assert.Equal(t, "-9.99", adj.DestinationAmount)
				assert.Equal(t, "EUR", adj.DestinationCurrency)
				assert.Equal(t, "audit correction", req.Notes)
				assert.Equal(t, []int32{99}, req.TagIds)
				require.NotNil(t, req.CategoryId)
				assert.Equal(t, int32(8), *req.CategoryId)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			txSvc := NewMockTransactionService(ctrl)
			txSvc.EXPECT().Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *transactionsv1.CreateTransactionRequest) (*transactionsv1.CreateTransactionResponse, error) {
					c.verify(t, req)
					return &transactionsv1.CreateTransactionResponse{
						Transaction: &gomoneypbv1.Transaction{Id: c.respID},
					}, nil
				})

			server := newTxServer(t, ctrl, txSvc)
			tool := server.MCPServer().GetTool("create_adjustment")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_adjustment",
					Arguments: c.args,
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "Transaction created with id")
		})
	}
}

func TestServer_HandleCreateAdjustment_Failure(t *testing.T) {
	baseArgs := func() map[string]any {
		return map[string]any{
			"transaction_date":       "2024-01-15T10:00:00Z",
			"title":                  "Balance fix",
			"destination_account_id": float64(5),
			"destination_amount":     "12.34",
			"destination_currency":   "USD",
		}
	}

	argsWith := func(overrides map[string]any, remove ...string) map[string]any {
		args := baseArgs()
		for k, v := range overrides {
			args[k] = v
		}
		for _, k := range remove {
			delete(args, k)
		}
		return args
	}

	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockTransactionService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing transaction_date",
			args:          argsWith(nil, "transaction_date"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "transaction_date is required",
		},
		{
			name:          "missing destination_account_id",
			args:          argsWith(nil, "destination_account_id"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "destination_account_id is required",
		},
		{
			name:          "missing destination_amount",
			args:          argsWith(nil, "destination_amount"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "destination_amount is required",
		},
		{
			name:          "invalid destination_amount",
			args:          argsWith(map[string]any{"destination_amount": "not-a-decimal"}),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "invalid destination_amount",
		},
		{
			name: "service returns error",
			args: baseArgs(),
			setupMock: func(m *MockTransactionService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db down"))
			},
			expectedError: "failed to create adjustment",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			txSvc := NewMockTransactionService(ctrl)
			c.setupMock(txSvc)

			server := newTxServer(t, ctrl, txSvc)
			tool := server.MCPServer().GetTool("create_adjustment")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_adjustment",
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

func TestServer_HandleUpdateExpense_Success(t *testing.T) {
	type tc struct {
		name   string
		args   map[string]any
		respID int64
		verify func(t *testing.T, req *transactionsv1.UpdateTransactionRequest)
	}

	cases := []tc{
		{
			name: "full fields",
			args: map[string]any{
				"id":                         float64(55),
				"transaction_date":           "2024-03-10T12:30:00Z",
				"title":                      "Updated Groceries",
				"source_account_id":          float64(10),
				"source_amount":              "-60.00",
				"source_currency":            "USD",
				"destination_account_id":     float64(20),
				"destination_amount":         "60.00",
				"destination_currency":       "USD",
				"fx_source_amount":           "-60.00",
				"fx_source_currency":         "USD",
				"notes":                      "updated note",
				"extra":                      map[string]any{"k": "v"},
				"tag_ids":                    []any{float64(4), float64(5)},
				"reference_number":           "REF-UPD",
				"internal_reference_numbers": []any{"INT-UPD"},
				"group_key":                  "grp-upd",
				"skip_rules":                 true,
				"category_id":                float64(77),
			},
			respID: 55,
			verify: func(t *testing.T, req *transactionsv1.UpdateTransactionRequest) {
				assert.Equal(t, int64(55), req.Id)
				require.NotNil(t, req.Transaction)
				assert.Equal(t, "Updated Groceries", req.Transaction.Title)
				exp := req.Transaction.GetExpense()
				require.NotNil(t, exp)
				assert.Equal(t, "-60.00", exp.SourceAmount)
				assert.Equal(t, int32(10), exp.SourceAccountId)
				assert.Equal(t, int32(20), exp.DestinationAccountId)
				require.NotNil(t, exp.FxSourceAmount)
				assert.Equal(t, "-60.00", *exp.FxSourceAmount)
				assert.Equal(t, []int32{4, 5}, req.Transaction.TagIds)
				require.NotNil(t, req.Transaction.CategoryId)
				assert.Equal(t, int32(77), *req.Transaction.CategoryId)
				assert.True(t, req.Transaction.SkipRules)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			txSvc := NewMockTransactionService(ctrl)
			txSvc.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *transactionsv1.UpdateTransactionRequest) (*transactionsv1.UpdateTransactionResponse, error) {
					c.verify(t, req)
					return &transactionsv1.UpdateTransactionResponse{
						Transaction: &gomoneypbv1.Transaction{Id: c.respID},
					}, nil
				})

			server := newTxServer(t, ctrl, txSvc)
			tool := server.MCPServer().GetTool("update_expense")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "update_expense",
					Arguments: c.args,
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "Transaction 55 updated")
		})
	}
}

func TestServer_HandleUpdateExpense_Failure(t *testing.T) {
	baseArgs := func() map[string]any {
		return map[string]any{
			"id":                     float64(1),
			"transaction_date":       "2024-01-15T10:00:00Z",
			"title":                  "Lunch",
			"source_account_id":      float64(1),
			"source_amount":          "-15.50",
			"source_currency":        "USD",
			"destination_account_id": float64(2),
			"destination_amount":     "15.50",
			"destination_currency":   "USD",
		}
	}

	argsWithout := func(remove ...string) map[string]any {
		args := baseArgs()
		for _, k := range remove {
			delete(args, k)
		}
		return args
	}

	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockTransactionService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing id",
			args:          argsWithout("id"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "id is required",
		},
		{
			name:          "missing transaction_date",
			args:          argsWithout("transaction_date"),
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "transaction_date is required",
		},
		{
			name: "service returns error",
			args: baseArgs(),
			setupMock: func(m *MockTransactionService) {
				m.EXPECT().Update(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))
			},
			expectedError: "failed to update expense",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			txSvc := NewMockTransactionService(ctrl)
			c.setupMock(txSvc)

			server := newTxServer(t, ctrl, txSvc)
			tool := server.MCPServer().GetTool("update_expense")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "update_expense",
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
