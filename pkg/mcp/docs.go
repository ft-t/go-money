package mcp

import (
	"embed"
	"io/fs"
	"strings"
)

//go:embed docs/*
var docsFS embed.FS

func GetContext() (string, error) {
	var sb strings.Builder

	err := fs.WalkDir(docsFS, "docs", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		content, readErr := docsFS.ReadFile(path)
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
