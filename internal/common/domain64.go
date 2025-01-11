package common

import (
	"encoding/json"
	"fmt"
	"strconv"

	"go.uber.org/zap"
)

type TLD64s map[string]TLD64

type TLD64 struct {
	Domain64ToFQDN map[string]string `json:"Domain64ToFQDN"`
	Domain64s      []string          `json:"Domain64s"`
	FQDNToDomain64 map[string]string `json:"FQDNToDomain64"`
	FQDNs          []string          `json:"FQDNs"`
	RFQDNs         []string          `json:"RFQDNs"`
}

type Domain64 struct {
	Hex string
	Dec int64
}

func NewTLD64s(bytes []byte) (TLD64s, error) {
	var tld TLD64s

	err := json.Unmarshal(bytes, &tld)
	if err != nil {
		return nil, fmt.Errorf("could not parse TLD64 JSON")
	}

	return tld, nil
}

func D64HexToDec(h string) int64 {
	value, err := strconv.ParseUint(h, 16, 64)
	if err != nil {
		zap.L().Error("could not convert Domain64",
			zap.String("Domain64", h))
	}

	//nolint:gosec
	return int64(value)
}

func D64DecToHex(n int64) string {
	return fmt.Sprintf("%016x", n)
}
