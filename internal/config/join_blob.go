package config

import (
	"net/url"
	"strings"
)

// JoinBlobDisplayURL appends path segments (percent-encoded per segment) to a GitLab blob base URL.
// base must be the resolved prefix ending before the file path (e.g. …/-/blob/main).
func JoinBlobDisplayURL(base, filePath string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	filePath = strings.TrimSpace(filePath)
	if base == "" || filePath == "" {
		return ""
	}
	parts := strings.Split(filePath, "/")
	var b strings.Builder
	b.WriteString(base)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		b.WriteByte('/')
		b.WriteString(url.PathEscape(p))
	}
	return b.String()
}
