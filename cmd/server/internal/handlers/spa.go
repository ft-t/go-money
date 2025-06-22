package handlers

import (
	"net/http"
	"os"
	"path/filepath"
)

func SpaHandler(staticDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(staticDir, r.URL.Path)

		_, err := os.Stat(path)
		if os.IsNotExist(err) || r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}

		http.FileServer(http.Dir(staticDir)).ServeHTTP(w, r)
	}
}
