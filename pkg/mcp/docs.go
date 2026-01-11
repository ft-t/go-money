package mcp

import (
	"io/fs"
	"os"
	"strings"
)

func ReadDocsFromPath(path string) (string, error) {
	var sb strings.Builder

	docFs := os.DirFS(path)
	err := fs.WalkDir(docFs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, readErr := fs.ReadFile(docFs, path)
		if readErr != nil {
			return readErr
		}

		sb.WriteString("# File: ")
		sb.WriteString(path)
		sb.WriteString("\n\n")
		sb.Write(content)
		sb.WriteString("\n\n---\n\n")

		return nil
	})

	if err != nil {
		return "", err
	}

	return sb.String(), nil
}
