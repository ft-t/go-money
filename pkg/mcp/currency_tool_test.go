package mcp_test

import (
	"context"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/golang/mock/gomock"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ft-t/go-money/pkg/currency"
	gomcp "github.com/ft-t/go-money/pkg/mcp"
	"github.com/ft-t/go-money/pkg/testingutils"
)

func newCurrencyServer(t *testing.T, ctrl *gomock.Controller, currencySvc *MockCurrencyConverterService) *gomcp.Server {
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
		CurrencySvc:    currencySvc,
	})
}

func TestServer_HandleConvertCurrency_Success(t *testing.T) {
	type tc struct {
		name     string
		args     map[string]any
		wantFrom string
		wantTo   string
		wantAmt  decimal.Decimal
		quote    *currency.Quote
		expected map[string]string
	}

	cases := []tc{
		{
			name: "standard cross-currency conversion",
			args: map[string]any{
				"from":   "USD",
				"to":     "EUR",
				"amount": "100",
			},
			wantFrom: "USD",
			wantTo:   "EUR",
			wantAmt:  decimal.NewFromInt(100),
			quote: &currency.Quote{
				From:         "USD",
				To:           "EUR",
				Amount:       decimal.NewFromInt(100),
				Converted:    decimal.RequireFromString("91.5"),
				FromRate:     decimal.NewFromInt(1),
				ToRate:       decimal.RequireFromString("0.915"),
				BaseCurrency: "USD",
			},
			expected: map[string]string{
				"from":          "USD",
				"to":            "EUR",
				"amount":        "100",
				"converted":     "91.5",
				"from_rate":     "1",
				"to_rate":       "0.915",
				"base_currency": "USD",
			},
		},
		{
			name: "same currency passthrough",
			args: map[string]any{
				"from":   "USD",
				"to":     "USD",
				"amount": "42.5",
			},
			wantFrom: "USD",
			wantTo:   "USD",
			wantAmt:  decimal.RequireFromString("42.5"),
			quote: &currency.Quote{
				From:         "USD",
				To:           "USD",
				Amount:       decimal.RequireFromString("42.5"),
				Converted:    decimal.RequireFromString("42.5"),
				FromRate:     decimal.NewFromInt(1),
				ToRate:       decimal.NewFromInt(1),
				BaseCurrency: "USD",
			},
			expected: map[string]string{
				"from":          "USD",
				"to":            "USD",
				"amount":        "42.5",
				"converted":     "42.5",
				"from_rate":     "1",
				"to_rate":       "1",
				"base_currency": "USD",
			},
		},
		{
			name: "large decimal with precision",
			args: map[string]any{
				"from":   "JPY",
				"to":     "USD",
				"amount": "1500000.123456",
			},
			wantFrom: "JPY",
			wantTo:   "USD",
			wantAmt:  decimal.RequireFromString("1500000.123456"),
			quote: &currency.Quote{
				From:         "JPY",
				To:           "USD",
				Amount:       decimal.RequireFromString("1500000.123456"),
				Converted:    decimal.RequireFromString("9876.54321"),
				FromRate:     decimal.RequireFromString("151.85"),
				ToRate:       decimal.NewFromInt(1),
				BaseCurrency: "USD",
			},
			expected: map[string]string{
				"from":          "JPY",
				"to":            "USD",
				"amount":        "1500000.123456",
				"converted":     "9876.54321",
				"from_rate":     "151.85",
				"to_rate":       "1",
				"base_currency": "USD",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			currencySvc := NewMockCurrencyConverterService(ctrl)
			currencySvc.EXPECT().Quote(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, from, to string, amount decimal.Decimal) (*currency.Quote, error) {
					assert.Equal(t, c.wantFrom, from)
					assert.Equal(t, c.wantTo, to)
					assert.True(t, c.wantAmt.Equal(amount), "amount mismatch: want %s got %s", c.wantAmt.String(), amount.String())
					return c.quote, nil
				})

			server := newCurrencyServer(t, ctrl, currencySvc)
			tool := server.MCPServer().GetTool("convert_currency")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "convert_currency",
					Arguments: c.args,
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			text := result.Content[0].(mcp.TextContent).Text
			for k, v := range c.expected {
				assert.Contains(t, text, "\""+k+"\"")
				assert.Contains(t, text, "\""+v+"\"")
			}
		})
	}
}

func TestServer_HandleConvertCurrency_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockCurrencyConverterService)
		expectedError string
	}

	cases := []tc{
		{
			name: "missing from",
			args: map[string]any{
				"to":     "EUR",
				"amount": "100",
			},
			setupMock:     func(m *MockCurrencyConverterService) {},
			expectedError: "from, to, and amount are required",
		},
		{
			name: "missing to",
			args: map[string]any{
				"from":   "USD",
				"amount": "100",
			},
			setupMock:     func(m *MockCurrencyConverterService) {},
			expectedError: "from, to, and amount are required",
		},
		{
			name: "missing amount",
			args: map[string]any{
				"from": "USD",
				"to":   "EUR",
			},
			setupMock:     func(m *MockCurrencyConverterService) {},
			expectedError: "from, to, and amount are required",
		},
		{
			name: "empty string from",
			args: map[string]any{
				"from":   "",
				"to":     "EUR",
				"amount": "100",
			},
			setupMock:     func(m *MockCurrencyConverterService) {},
			expectedError: "from, to, and amount are required",
		},
		{
			name: "invalid amount",
			args: map[string]any{
				"from":   "USD",
				"to":     "EUR",
				"amount": "not-a-decimal",
			},
			setupMock:     func(m *MockCurrencyConverterService) {},
			expectedError: "invalid amount",
		},
		{
			name: "service returns error",
			args: map[string]any{
				"from":   "USD",
				"to":     "EUR",
				"amount": "100",
			},
			setupMock: func(m *MockCurrencyConverterService) {
				m.EXPECT().Quote(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errors.New("rate unavailable"))
			},
			expectedError: "failed to convert",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			currencySvc := NewMockCurrencyConverterService(ctrl)
			c.setupMock(currencySvc)

			server := newCurrencyServer(t, ctrl, currencySvc)
			tool := server.MCPServer().GetTool("convert_currency")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "convert_currency",
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
