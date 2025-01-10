package util

import (
	// SHA1 is a weak cypher. We could use sha256.
	// However, we're not using it for crypto purposes,
	// just for generating a temporary name. We do not
	// expect collisions.
	//nolint:gosec
	"crypto/sha1"
	"fmt"
	"strings"
)

type Scheme int

const (
	HTTP Scheme = iota
	HTTPS
)

func (s Scheme) String() string {
	return [...]string{"http", "https"}[s]
}

func ToScheme(scheme string) Scheme {
	switch strings.ToLower(scheme) {
	case "http":
		return 0
	case "https":
		return 1
	default:
		return 1
	}
}

func (s Scheme) EnumIndex() int {
	return int(s)
}

type Extension int

const (
	JSON Extension = iota
	Raw
)

func (e Extension) String() string {
	return [...]string{"json", "raw"}[e]
}

func (e Extension) EnumIndex() int {
	return int(e)
}

type Key struct {
	Scheme    Scheme
	Host      string
	Path      string
	Extension Extension
}

//nolint:gosec
func (k *Key) SHA1() string {
	sha := fmt.Sprintf("%x", sha1.Sum([]byte(k.Host+k.Path)))

	return sha
}

func (k *Key) Render() string {
	return fmt.Sprintf("%s/%s.%s", k.Host, k.SHA1(), k.Extension)
}

func CreateS3Key(scheme Scheme, host string, path string, ext Extension) *Key {
	return &Key{
		Scheme:    scheme,
		Host:      host,
		Path:      path,
		Extension: ext,
	}
}

func (k *Key) Copy() *Key {
	return &Key{
		Scheme:    k.Scheme,
		Host:      k.Host,
		Path:      k.Path,
		Extension: k.Extension,
	}
}
