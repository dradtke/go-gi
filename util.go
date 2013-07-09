package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

func CamelCase(str string) string {
	var b bytes.Buffer
	parts := strings.Split(str, "_")
	for _, part := range parts {
		b.WriteString(strings.Title(part))
	}
	return b.String()
}

func Search(path, filename string) string {
	for _, dir := range strings.Split(path, string(os.PathListSeparator)) {
		f := filepath.Join(dir, filename)
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			return f
		}
	}
	return ""
}
