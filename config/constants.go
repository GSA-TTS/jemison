package config

import (
	"embed"
	"fmt"

	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

//go:embed constants.json
var ConstFS embed.FS

/*
Load the bytes into RAM, and leave them there.
Assume over the live of a service we'll hit
this file a whole bunch of times. And, it never
changes during a single deploy, so... :shrug:.
*/
var cachedConstants []byte

func primeConstants() {
	// Cache this
	if cachedConstants == nil {
		bytes, err := ConstFS.ReadFile("constants.json")
		if err != nil {
			zap.L().Fatal("could not read constants from embedded FS")
		}

		cachedConstants = bytes
	}
}

func GetScheme(scheme string) int {
	primeConstants()

	v := gjson.GetBytes(cachedConstants, "SchemeToConst."+scheme).Int()

	return int(v)
}

func GetContentType(ct string) int {
	primeConstants()

	v := gjson.GetBytes(cachedConstants, "ContentTypeToConst."+ct).Int()

	return int(v)
}

func GetTLD(tld string) int {
	primeConstants()

	v := gjson.GetBytes(cachedConstants, "TldToConst."+tld).Int()

	return int(v)
}

func IntToTld(i int) string {
	primeConstants()

	search_string := "ConstToTld." + fmt.Sprintf("%x", i)
	v := gjson.GetBytes(cachedConstants, search_string).String()

	return v
}
