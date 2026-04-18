package mcp_test

import (
	"context"
	"fmt"
	"testing"

	tagsv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/tags/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/golang/mock/gomock"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gomcp "github.com/ft-t/go-money/pkg/mcp"
	"github.com/ft-t/go-money/pkg/testingutils"
	"github.com/ft-t/go-money/pkg/transactions"
)

func TestServer_HandleListTags_Success(t *testing.T) {
	type tc struct {
		name     string
		tags     []*tagsv1.ListTagsResponse_TagItem
		expected string
	}

	cases := []tc{
		{
			name:     "no tags",
			tags:     nil,
			expected: "No tags found",
		},
		{
			name: "single tag",
			tags: []*tagsv1.ListTagsResponse_TagItem{
				{Tag: &gomoneypbv1.Tag{Id: 1, Name: "Food", Color: "#ff0000", Icon: "utensils"}},
			},
			expected: "Food",
		},
		{
			name: "multiple tags",
			tags: []*tagsv1.ListTagsResponse_TagItem{
				{Tag: &gomoneypbv1.Tag{Id: 1, Name: "Food", Color: "#ff0000", Icon: "utensils"}},
				{Tag: &gomoneypbv1.Tag{Id: 2, Name: "Travel", Color: "#00ff00", Icon: "plane"}},
			},
			expected: "Travel",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			tagsSvc := NewMockTagsService(ctrl)
			tagsSvc.EXPECT().ListTags(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *tagsv1.ListTagsRequest) (*tagsv1.ListTagsResponse, error) {
					assert.NotNil(t, req)
					return &tagsv1.ListTagsResponse{Tags: c.tags}, nil
				})

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: NewMockCategoryService(ctrl),
				RulesSvc:    NewMockRulesService(ctrl),
				DryRunSvc:   NewMockDryRunService(ctrl),
				TagsSvc:     tagsSvc,
			})

			tool := server.MCPServer().GetTool("list_tags")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "list_tags",
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expected)
		})
	}
}

func TestServer_HandleListTags_Failure(t *testing.T) {
	type tc struct {
		name          string
		setupMock     func(*MockTagsService)
		expectedError string
	}

	cases := []tc{
		{
			name: "service error",
			setupMock: func(m *MockTagsService) {
				m.EXPECT().ListTags(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("db connection failed"))
			},
			expectedError: "failed to list tags",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			tagsSvc := NewMockTagsService(ctrl)
			c.setupMock(tagsSvc)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: NewMockCategoryService(ctrl),
				RulesSvc:    NewMockRulesService(ctrl),
				DryRunSvc:   NewMockDryRunService(ctrl),
				TagsSvc:     tagsSvc,
			})

			tool := server.MCPServer().GetTool("list_tags")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "list_tags",
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expectedError)
		})
	}
}

func TestServer_HandleCreateTag_Success(t *testing.T) {
	type tc struct {
		name     string
		tagName  string
		color    string
		icon     string
		args     map[string]any
		respID   int32
		expected string
	}

	cases := []tc{
		{
			name:     "create tag with all fields",
			tagName:  "Food",
			color:    "#ff0000",
			icon:     "utensils",
			args:     map[string]any{"name": "Food", "color": "#ff0000", "icon": "utensils"},
			respID:   1,
			expected: "Tag created with id 1",
		},
		{
			name:     "create tag name only",
			tagName:  "Travel",
			color:    "",
			icon:     "",
			args:     map[string]any{"name": "Travel"},
			respID:   2,
			expected: "Tag created with id 2",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			tagsSvc := NewMockTagsService(ctrl)
			tagsSvc.EXPECT().CreateTag(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *tagsv1.CreateTagRequest) (*tagsv1.CreateTagResponse, error) {
					assert.Equal(t, c.tagName, req.Name)
					assert.Equal(t, c.color, req.Color)
					assert.Equal(t, c.icon, req.Icon)
					return &tagsv1.CreateTagResponse{
						Tag: &gomoneypbv1.Tag{
							Id:    c.respID,
							Name:  c.tagName,
							Color: c.color,
							Icon:  c.icon,
						},
					}, nil
				})

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: NewMockCategoryService(ctrl),
				RulesSvc:    NewMockRulesService(ctrl),
				DryRunSvc:   NewMockDryRunService(ctrl),
				TagsSvc:     tagsSvc,
			})

			tool := server.MCPServer().GetTool("create_tag")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_tag",
					Arguments: c.args,
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expected)
		})
	}
}

func TestServer_HandleCreateTag_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockTagsService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing name",
			args:          map[string]any{},
			setupMock:     func(m *MockTagsService) {},
			expectedError: "name parameter is required",
		},
		{
			name:          "empty name",
			args:          map[string]any{"name": ""},
			setupMock:     func(m *MockTagsService) {},
			expectedError: "name parameter is required",
		},
		{
			name: "service error",
			args: map[string]any{"name": "Test"},
			setupMock: func(m *MockTagsService) {
				m.EXPECT().CreateTag(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("duplicate name"))
			},
			expectedError: "failed to create tag",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			tagsSvc := NewMockTagsService(ctrl)
			c.setupMock(tagsSvc)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: NewMockCategoryService(ctrl),
				RulesSvc:    NewMockRulesService(ctrl),
				DryRunSvc:   NewMockDryRunService(ctrl),
				TagsSvc:     tagsSvc,
			})

			tool := server.MCPServer().GetTool("create_tag")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_tag",
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

func TestServer_HandleUpdateTag_Success(t *testing.T) {
	type tc struct {
		name     string
		tagID    float64
		tagName  string
		color    string
		icon     string
		args     map[string]any
		expected string
	}

	cases := []tc{
		{
			name:     "update tag with all fields",
			tagID:    1,
			tagName:  "Updated Food",
			color:    "#00ff00",
			icon:     "bowl",
			args:     map[string]any{"id": float64(1), "name": "Updated Food", "color": "#00ff00", "icon": "bowl"},
			expected: "Tag 1 updated",
		},
		{
			name:     "update tag name only",
			tagID:    5,
			tagName:  "Renamed",
			color:    "",
			icon:     "",
			args:     map[string]any{"id": float64(5), "name": "Renamed"},
			expected: "Tag 5 updated",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			tagsSvc := NewMockTagsService(ctrl)
			tagsSvc.EXPECT().UpdateTag(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *tagsv1.UpdateTagRequest) (*tagsv1.UpdateTagResponse, error) {
					assert.Equal(t, int32(c.tagID), req.Id)
					assert.Equal(t, c.tagName, req.Name)
					assert.Equal(t, c.color, req.Color)
					assert.Equal(t, c.icon, req.Icon)
					return &tagsv1.UpdateTagResponse{
						Tag: &gomoneypbv1.Tag{
							Id:    int32(c.tagID),
							Name:  c.tagName,
							Color: c.color,
							Icon:  c.icon,
						},
					}, nil
				})

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: NewMockCategoryService(ctrl),
				RulesSvc:    NewMockRulesService(ctrl),
				DryRunSvc:   NewMockDryRunService(ctrl),
				TagsSvc:     tagsSvc,
			})

			tool := server.MCPServer().GetTool("update_tag")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "update_tag",
					Arguments: c.args,
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expected)
		})
	}
}

func TestServer_HandleUpdateTag_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockTagsService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing id",
			args:          map[string]any{"name": "Test"},
			setupMock:     func(m *MockTagsService) {},
			expectedError: "id parameter is required",
		},
		{
			name:          "missing name",
			args:          map[string]any{"id": float64(1)},
			setupMock:     func(m *MockTagsService) {},
			expectedError: "name parameter is required",
		},
		{
			name: "service error",
			args: map[string]any{"id": float64(999), "name": "Test"},
			setupMock: func(m *MockTagsService) {
				m.EXPECT().UpdateTag(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))
			},
			expectedError: "failed to update tag",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			tagsSvc := NewMockTagsService(ctrl)
			c.setupMock(tagsSvc)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: NewMockCategoryService(ctrl),
				RulesSvc:    NewMockRulesService(ctrl),
				DryRunSvc:   NewMockDryRunService(ctrl),
				TagsSvc:     tagsSvc,
			})

			tool := server.MCPServer().GetTool("update_tag")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "update_tag",
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

func TestServer_HandleDeleteTag_Success(t *testing.T) {
	type tc struct {
		name     string
		tagID    float64
		expected string
	}

	cases := []tc{
		{
			name:     "delete tag",
			tagID:    1,
			expected: "Tag 1 deleted",
		},
		{
			name:     "delete another tag",
			tagID:    42,
			expected: "Tag 42 deleted",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			tagsSvc := NewMockTagsService(ctrl)
			tagsSvc.EXPECT().DeleteTag(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *tagsv1.DeleteTagRequest) error {
					assert.Equal(t, int32(c.tagID), req.Id)
					return nil
				})

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: NewMockCategoryService(ctrl),
				RulesSvc:    NewMockRulesService(ctrl),
				DryRunSvc:   NewMockDryRunService(ctrl),
				TagsSvc:     tagsSvc,
			})

			tool := server.MCPServer().GetTool("delete_tag")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "delete_tag",
					Arguments: map[string]any{"id": c.tagID},
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expected)
		})
	}
}

func TestServer_HandleDeleteTag_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockTagsService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing id",
			args:          map[string]any{},
			setupMock:     func(m *MockTagsService) {},
			expectedError: "id parameter is required",
		},
		{
			name: "service error",
			args: map[string]any{"id": float64(1)},
			setupMock: func(m *MockTagsService) {
				m.EXPECT().DeleteTag(gomock.Any(), gomock.Any()).
					Return(errors.New("not found"))
			},
			expectedError: "failed to delete tag",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			tagsSvc := NewMockTagsService(ctrl)
			c.setupMock(tagsSvc)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: NewMockCategoryService(ctrl),
				RulesSvc:    NewMockRulesService(ctrl),
				DryRunSvc:   NewMockDryRunService(ctrl),
				TagsSvc:     tagsSvc,
			})

			tool := server.MCPServer().GetTool("delete_tag")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "delete_tag",
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

func TestServer_HandleBulkSetTransactionTags_Success(t *testing.T) {
	type tc struct {
		name        string
		assignments []map[string]any
		expected    []transactions.TagsAssignment
	}

	cases := []tc{
		{
			name: "set single tag list",
			assignments: []map[string]any{
				{"transaction_id": float64(1), "tag_ids": []any{float64(5)}},
			},
			expected: []transactions.TagsAssignment{
				{TransactionID: 1, TagIDs: []int32{5}},
			},
		},
		{
			name: "set multiple assignments",
			assignments: []map[string]any{
				{"transaction_id": float64(1), "tag_ids": []any{float64(5), float64(10)}},
				{"transaction_id": float64(2), "tag_ids": []any{float64(7)}},
			},
			expected: []transactions.TagsAssignment{
				{TransactionID: 1, TagIDs: []int32{5, 10}},
				{TransactionID: 2, TagIDs: []int32{7}},
			},
		},
		{
			name: "clear tags with empty array",
			assignments: []map[string]any{
				{"transaction_id": float64(1), "tag_ids": []any{}},
			},
			expected: []transactions.TagsAssignment{
				{TransactionID: 1, TagIDs: []int32{}},
			},
		},
		{
			name: "mixed set and clear",
			assignments: []map[string]any{
				{"transaction_id": float64(1), "tag_ids": []any{float64(5), float64(10)}},
				{"transaction_id": float64(2), "tag_ids": []any{}},
			},
			expected: []transactions.TagsAssignment{
				{TransactionID: 1, TagIDs: []int32{5, 10}},
				{TransactionID: 2, TagIDs: []int32{}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			txSvc := NewMockTransactionService(ctrl)
			expected := c.expected
			txSvc.EXPECT().BulkSetTags(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, got []transactions.TagsAssignment) error {
					assert.Equal(t, expected, got)
					return nil
				})

			server := newTxServer(t, ctrl, txSvc)
			tool := server.MCPServer().GetTool("bulk_set_transaction_tags")
			require.NotNil(t, tool)

			assignmentsAny := make([]any, len(c.assignments))
			for i, a := range c.assignments {
				assignmentsAny[i] = a
			}

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "bulk_set_transaction_tags",
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

func TestServer_HandleBulkSetTransactionTags_Failure(t *testing.T) {
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
				map[string]any{"tag_ids": []any{float64(5)}},
			}},
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "assignment[0].transaction_id is required",
		},
		{
			name: "missing tag_ids",
			args: map[string]any{"assignments": []any{
				map[string]any{"transaction_id": float64(1)},
			}},
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "assignment[0].tag_ids is required",
		},
		{
			name: "tag_ids not an array",
			args: map[string]any{"assignments": []any{
				map[string]any{"transaction_id": float64(1), "tag_ids": "not-an-array"},
			}},
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "assignment[0].tag_ids is required and must be an array",
		},
		{
			name: "tag_ids element not a number",
			args: map[string]any{"assignments": []any{
				map[string]any{"transaction_id": float64(1), "tag_ids": []any{"invalid"}},
			}},
			setupMock:     func(m *MockTransactionService) {},
			expectedError: "assignment[0].tag_ids[0] must be a number",
		},
		{
			name: "service returns error",
			args: map[string]any{"assignments": []any{
				map[string]any{"transaction_id": float64(1), "tag_ids": []any{float64(5)}},
			}},
			setupMock: func(m *MockTransactionService) {
				m.EXPECT().BulkSetTags(gomock.Any(), gomock.Any()).
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
			tool := server.MCPServer().GetTool("bulk_set_transaction_tags")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "bulk_set_transaction_tags",
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
