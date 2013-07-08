package gogi

import (
	"bytes"
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
