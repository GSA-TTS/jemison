//nolint:testpackage
package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConstant(t *testing.T) {
	assert.Equal(t, 1, GetScheme("https"))
	assert.Equal(t, 2, GetScheme("http"))
	assert.Equal(t, 1, GetTLD("gov"))
}
