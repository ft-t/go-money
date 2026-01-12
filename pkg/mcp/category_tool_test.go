package mcp_test

import (
	"context"
	"testing"

	categoriesv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/categories/v1"
	gomoneypbv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/v1"
	"github.com/cockroachdb/errors"
	"github.com/golang/mock/gomock"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gomcp "github.com/ft-t/go-money/pkg/mcp"
	"github.com/ft-t/go-money/pkg/testingutils"
)

func TestServer_HandleCreateCategory_Success(t *testing.T) {
	type tc struct {
		name     string
		catName  string
		respID   int32
		expected string
	}

	cases := []tc{
		{
			name:     "create category",
			catName:  "Food",
			respID:   1,
			expected: "Category created with id 1",
		},
		{
			name:     "create category with special chars",
			catName:  "Groceries & Shopping",
			respID:   2,
			expected: "Category created with id 2",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			catSvc := NewMockCategoryService(ctrl)
			catSvc.EXPECT().CreateCategory(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *categoriesv1.CreateCategoryRequest) (*categoriesv1.CreateCategoryResponse, error) {
					assert.Equal(t, c.catName, req.Category.Name)
					return &categoriesv1.CreateCategoryResponse{
						Category: &gomoneypbv1.Category{
							Id:   c.respID,
							Name: c.catName,
						},
					}, nil
				})

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("create_category")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_category",
					Arguments: map[string]any{"name": c.catName},
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expected)
		})
	}
}

func TestServer_HandleCreateCategory_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockCategoryService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing name",
			args:          map[string]any{},
			setupMock:     func(m *MockCategoryService) {},
			expectedError: "name parameter is required",
		},
		{
			name:          "empty name",
			args:          map[string]any{"name": ""},
			setupMock:     func(m *MockCategoryService) {},
			expectedError: "name parameter is required",
		},
		{
			name: "service error",
			args: map[string]any{"name": "Test"},
			setupMock: func(m *MockCategoryService) {
				m.EXPECT().CreateCategory(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("duplicate name"))
			},
			expectedError: "failed to create category",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			catSvc := NewMockCategoryService(ctrl)
			c.setupMock(catSvc)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("create_category")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "create_category",
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

func TestServer_HandleUpdateCategory_Success(t *testing.T) {
	type tc struct {
		name     string
		catID    float64
		catName  string
		expected string
	}

	cases := []tc{
		{
			name:     "update category",
			catID:    1,
			catName:  "Updated Food",
			expected: "Category 1 updated",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			catSvc := NewMockCategoryService(ctrl)
			catSvc.EXPECT().UpdateCategory(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, req *categoriesv1.UpdateCategoryRequest) (*categoriesv1.UpdateCategoryResponse, error) {
					assert.Equal(t, int32(c.catID), req.Category.Id)
					assert.Equal(t, c.catName, req.Category.Name)
					return &categoriesv1.UpdateCategoryResponse{
						Category: &gomoneypbv1.Category{
							Id:   int32(c.catID),
							Name: c.catName,
						},
					}, nil
				})

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("update_category")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "update_category",
					Arguments: map[string]any{"id": c.catID, "name": c.catName},
				},
			})

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.False(t, result.IsError)
			assert.Contains(t, result.Content[0].(mcp.TextContent).Text, c.expected)
		})
	}
}

func TestServer_HandleUpdateCategory_Failure(t *testing.T) {
	type tc struct {
		name          string
		args          map[string]any
		setupMock     func(*MockCategoryService)
		expectedError string
	}

	cases := []tc{
		{
			name:          "missing id",
			args:          map[string]any{"name": "Test"},
			setupMock:     func(m *MockCategoryService) {},
			expectedError: "id parameter is required",
		},
		{
			name:          "missing name",
			args:          map[string]any{"id": float64(1)},
			setupMock:     func(m *MockCategoryService) {},
			expectedError: "name parameter is required",
		},
		{
			name: "service error",
			args: map[string]any{"id": float64(999), "name": "Test"},
			setupMock: func(m *MockCategoryService) {
				m.EXPECT().UpdateCategory(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("not found"))
			},
			expectedError: "failed to update category",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			gormDB, mockDB, _ := testingutils.GormMock()
			defer func() { _ = mockDB.Close() }()

			catSvc := NewMockCategoryService(ctrl)
			c.setupMock(catSvc)

			server := gomcp.NewServer(&gomcp.ServerConfig{
				DB:          gormDB,
				Docs:        "test docs",
				CategorySvc: catSvc,
			})

			mcpServer := server.MCPServer()
			tool := mcpServer.GetTool("update_category")
			require.NotNil(t, tool)

			result, err := tool.Handler(context.Background(), mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      "update_category",
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
