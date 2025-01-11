//nolint:all
package config

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFQDN2Domain64(t *testing.T) {
	// Test a lookup against the decimal value
	v, err := FQDNToDomain64("my.goarmy.com")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(216172790703718656), v)

	// Test the same lookup against a hex parse
	v, err = FQDNToDomain64("my.goarmy.com")
	if err != nil {
		t.Error(err)
	}
	i, err := strconv.ParseInt("0300000200000100", 16, 64)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(i), v)

	// And, something in a different TLD
	v, err = FQDNToDomain64("product-guide.18f.gov")
	if err != nil {
		t.Error(err)
	}
	i, err = strconv.ParseInt("0100000100000600", 16, 64)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, int64(i), v)

}

func TestIAmAChildlessDomain(t *testing.T) {
	v, err := FQDNToDomain64("cloud.gov")
	if err != nil {
		t.Error(err)
	}
	// 0100002400000000
	assert.Equal(t, int64(72057748656750592), v)

}

func TestGetAllFQDN(t *testing.T) {
	all := GetAllFQDNToDomain64()
	assert.Greater(t, len(all), 1000)
}

func TestEnoughHosts(t *testing.T) {
	three := GetListOfHosts("three")
	assert.Equal(t, 3, len(three))
}

func TestDomain64ToFQDN(t *testing.T) {
	v, err := Domain64ToFQDN(72057748656750592)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "cloud.gov", v)

	v, err = Domain64ToFQDN(216172790703718656)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "my.goarmy.com", v)

	v, err = Domain64ToFQDN(72057911865508096)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "www.fac.gov", v)

}
