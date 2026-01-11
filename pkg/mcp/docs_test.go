package mcp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ft-t/go-money/pkg/mcp"
)

func TestReadDocsFromPath_Success(t *testing.T) {
	type tc struct {
		name     string
		files    map[string]string
		expected []string
	}

	cases := []tc{
		{
			name: "single markdown file",
			files: map[string]string{
				"test.md": "# Test Content",
			},
			expected: []string{"# File: test.md", "# Test Content"},
		},
		{
			name: "multiple markdown files",
			files: map[string]string{
				"first.md":  "First content",
				"second.md": "Second content",
			},
			expected: []string{"# File: first.md", "First content", "# File: second.md", "Second content"},
		},
		{
			name: "nested directories",
			files: map[string]string{
				"root.md":           "Root content",
				"subdir/nested.md":  "Nested content",
				"subdir/another.md": "Another nested",
			},
			expected: []string{"# File: root.md", "Root content", "# File: subdir/nested.md", "Nested content"},
		},
		{
			name: "ignores non-markdown files",
			files: map[string]string{
				"doc.md":  "Markdown content",
				"data.txt": "Text content",
				"code.go":  "Go content",
			},
			expected: []string{"# File: doc.md", "Markdown content"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			for path, content := range c.files {
				fullPath := filepath.Join(tmpDir, path)
				require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
				require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
			}

			result, err := mcp.ReadDocsFromPath(tmpDir)

			require.NoError(t, err)
			for _, exp := range c.expected {
				assert.Contains(t, result, exp)
			}
		})
	}
}

func TestReadDocsFromPath_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := mcp.ReadDocsFromPath(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestReadDocsFromPath_Failure(t *testing.T) {
	type tc struct {
		name string
		path string
	}

	cases := []tc{
		{
			name: "non-existent directory",
			path: "/non/existent/path/that/does/not/exist",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := mcp.ReadDocsFromPath(c.path)

			assert.Error(t, err)
			assert.Empty(t, result)
		})
	}
}
