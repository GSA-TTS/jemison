package util

import (
	"net/url"
	"regexp"
	"strings"
)

type MimeType int

const (
	HTML MimeType = iota
	Plain
	PDF
	SQLite3
)

func (mt MimeType) String() string {
	return [...]string{
		"text/html",
		"text/plain",
		"application/pdf",
		"application/vnd.sqlite3",
	}[mt]
}

func (mt MimeType) EnumIndex() int {
	return int(mt)
}

func AtoZOnly(s string) string {
	var result strings.Builder

	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') {
			result.WriteByte(b)
		}
	}

	return result.String()
}

var mime_types = []string{
	"text/html",
	"text/plain",
	"application/pdf",
	"application/vnd.sqlite3",
}

func CleanMimeType(mime string) string {
	for _, m := range mime_types {
		if strings.Contains(mime, m) {
			return m
		}
	}

	// The unknown mime type
	return "application/octet-stream"
}

func GetMimeType(path string) string {
	m := map[string]string{
		"html":    "text/html",
		"htm":     "text/html",
		"json":    "application/json",
		"md":      "text/plain",
		"plain":   "text/plain",
		"pdf":     "application/pdf",
		"sqlite":  "application/vnd.sqlite3",
		"sqlite3": "application/vnd.sqlite3",
		"txt":     "text/plain",
		// https://www.iana.org/assignments/media-types/application/zstd
		"zstd": "application/zstd",
	}
	for tag, mime_type := range m {
		if strings.HasSuffix(path, tag) {
			return mime_type
		}
	}

	return m["json"]
}

func IsSearchableMimeType(mime string) bool {
	for _, m := range mime_types {
		if strings.Contains(mime, m) {
			return true
		}
	}

	return false
}

func CollapseWhitespace(s string) string {
	re := regexp.MustCompile(`\s\s+`)

	s = strings.TrimSpace(s)

	return re.ReplaceAllString(s, " ")
}

func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]

		return s
	} else {
		return s
	}
}

func CanonicalizeURL(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		//nolint:wrapcheck
		return "", err
	}

	u.Host = strings.ToLower(u.Host)
	if len(u.Path) > 1 {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}

	return u.Host + u.Path, nil
}
