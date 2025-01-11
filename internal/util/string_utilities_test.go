//nolint:testpackage
package util

import (
	"testing"

	"github.com/zeebo/assert"
)

func TestRemoveTrailingSlash(t *testing.T) {
	s := "/policies/"
	cleaned := TrimSuffix(s, "/")
	assert.Equal(t, "/policies", cleaned)
}
