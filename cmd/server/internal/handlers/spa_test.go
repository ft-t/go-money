package handlers_test

import (
	"github.com/ft-t/go-money/cmd/server/internal/handlers"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupStaticDir(t *testing.T) string {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.html"), []byte("INDEX"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte("FILE"), 0644))
	return dir
}

func TestSpaHandler_IndexHtmlOnRoot(t *testing.T) {
	dir := setupStaticDir(t)
	handler := handlers.SpaHandler(dir)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body := w.Body.String()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, body, "INDEX")
}

func TestSpaHandler_IndexHtmlOnNotExist(t *testing.T) {
	dir := setupStaticDir(t)
	handler := handlers.SpaHandler(dir)

	req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body := w.Body.String()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, body, "INDEX")
}

func TestSpaHandler_ServeStaticFile(t *testing.T) {
	dir := setupStaticDir(t)
	handler := handlers.SpaHandler(dir)

	req := httptest.NewRequest(http.MethodGet, "/file.txt", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body := w.Body.String()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, body, "FILE")
}

func TestSpaHandler_ForbiddenPath(t *testing.T) {
	dir := setupStaticDir(t)
	handler := handlers.SpaHandler(dir)

	req := httptest.NewRequest(http.MethodGet, "/../spa.go", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body := w.Body.String()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Contains(t, body, "invalid URL path")
}
