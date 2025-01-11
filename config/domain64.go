package config

import (
	"embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

type Schedule int

const (
	Daily Schedule = iota
	Weekly
	BiWeekly
	Monthly
	Quarterly
	BiAnnually
	Annually
	Default
)

//go:embed domain64/domain64.json
var Domain64FS embed.FS

/*
Load the bytes into RAM, and leave them there.
Assume over the live of a service we'll hit
this file a whole bunch of times. And, it never
changes during a single deploy, so... :shrug:.
*/
var cached_file []byte

func primeCache() {
	// Cache this
	if cached_file == nil {
		bytes, _ := Domain64FS.ReadFile("domain64/domain64.json")
		cached_file = bytes
	}
}

func tldAndEscaped(fqdn string) (string, string, error) {
	pieces := strings.Split(fqdn, ".")
	if len(pieces) < 2 {
		return "", "", fmt.Errorf("fqdn is too short: %s", fqdn)
	}

	tld := pieces[len(pieces)-1]
	// Escape the FQDN dots so it can be used with GJSON
	fqdn_as_json_key := strings.Replace(fqdn, ".", `\.`, -1)

	return tld, fqdn_as_json_key, nil
}

func FQDNToDomain64(fqdn string) (int64, error) {
	primeCache()

	tld, escaped, err := tldAndEscaped(fqdn)
	if err != nil {
		return 0, err
	}

	hex := gjson.GetBytes(cached_file, tld+".FQDNToDomain64."+escaped).String()

	value, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		//nolint:wrapcheck
		return 0, err
	}

	return int64(value), nil
}

func Domain64ToFQDN(domain64 int64) (string, error) {
	primeCache()

	h := fmt.Sprintf("%016X", domain64)
	v, _ := strconv.ParseInt(h[0:2], 16, 32)
	tld := IntToTld(int(v))
	fqdn := gjson.GetBytes(cached_file, tld+".Domain64ToFQDN."+h).String()
	// zap.L().Debug("d64tofqdn",
	// 	zap.String("h", h), zap.Int64("v", v), zap.String("tld", tld), zap.String("fqdn", fqdn))
	// log.Println("h", h, "v", v, "tld", tld, "fqdn", fqdn)
	return fqdn, nil
}

func RDomainToDomain64(rdomain string) string {
	primeCache()

	tld := strings.Split(rdomain, ".")[0]
	hex := gjson.GetBytes(cached_file, tld+".RDomainToDomain64."+strings.Replace(rdomain, ".", `\.`, -1)).String()

	return hex
}

func GetAllFQDNToDomain64() map[string]int64 {
	primeCache()

	tlds := gjson.GetBytes(cached_file, "TLDs").Array()
	all := make(map[string]int64)

	for _, tld := range tlds {
		m := gjson.GetBytes(cached_file, tld.String()+".FQDNToDomain64").Map()
		for fq, d64 := range m {
			dec, err := HexToDec64(d64.String())
			if err != nil {
				zap.L().Error("could not get decimal value for Domain64",
					zap.String("domain64", d64.String()), zap.String("fqdn", fq))
			}

			all[fq] = dec
		}
	}

	return all
}

func HexToDec64(hex string) (int64, error) {
	value, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		//nolint:wrapcheck
		return 0, err
	}

	return value, nil
}

// WARNING: This conevrsion makes some assumptions...
func Dec64ToHex(dec int64) string {
	return fmt.Sprintf("%016X", dec)
}

func GetSchedule(fqdn string) Schedule {
	primeCache()

	tld, escaped, err := tldAndEscaped(fqdn)
	hex := gjson.GetBytes(cached_file, tld+".FQDNToDomain64."+escaped).String()
	schedule := gjson.GetBytes(cached_file, tld+".Schedule."+hex).String()

	if err != nil {
		return Default
	} else {
		switch schedule {
		case "daily":
			return Daily
		case "weekly":
			return Weekly
		case "biweekly":
			return BiWeekly
		case "monthly":
			return Monthly
		case "Quarterly":
			return Quarterly
		case "BiAnnually":
			return BiAnnually
		case "Annually":
			return Annually
		}
	}

	return Default
}
