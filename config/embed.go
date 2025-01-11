package config

import (
	"bytes"
	"embed"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

//go:embed *.yaml
var ConfigFs embed.FS

func ReadConfigJsonnet(sonnetFilename string) string {
	bytes, _ := ConfigFs.ReadFile(sonnetFilename)

	vm := jsonnet.MakeVM()

	json, err := vm.EvaluateAnonymousSnippet(sonnetFilename, string(bytes))
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	return json
}

func ReadJsonConfig(jsonFilename string) string {
	json_bytes, err := ConfigFs.ReadFile(jsonFilename)
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	return string(json_bytes)
}

func GetYamlFileReader(yamlFilename string) *bytes.Reader {
	yaml_bytes, err := ConfigFs.ReadFile(yamlFilename)
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	return bytes.NewReader(yaml_bytes)
}

func GetListOfHosts(allowed_hosts string) []string {
	zap.L().Debug("reading in hosts", zap.String("allowed_hosts", allowed_hosts))

	cfg := ReadJsonConfig("allowed_hosts.yaml")

	// The variable `allowed_hosts` will be the key into the doc that has
	// a list of pairs. Each pair is a range of values, which tells us how
	// to filter the FQDN/D64 values.
	ranges := gjson.Get(cfg, allowed_hosts).Array()
	hosts := make([]string, 0)
	set := make(map[string]bool)

	all := GetAllFQDNToDomain64()

	for _, pair := range ranges {
		low := (pair.Array())[0].Int()
		high := (pair.Array())[1].Int()

		zap.L().Info("checking range", zap.Int64("low", low), zap.Int64("high", high))

		for fqdn, d64 := range all {
			if (d64 >= low) && (d64 <= high) {
				set[fqdn] = true
			}
		}
	}

	for fqdn := range set {
		hosts = append(hosts, fqdn)
	}

	return hosts
}

func GetHostBackend(host, schedule string) string {
	cfg := ReadJsonConfig(schedule)
	backend := "postgres"

	for _, section := range gjson.Parse(cfg).Get("@keys").Array() {
		for _, site := range gjson.Get(cfg, section.String()).Array() {
			if host == site.Get("host").String() {
				b := site.Get("backend").String()
				if b == "sqlite" || b == "postgres" {
					backend = b
				}
			}
		}
	}

	return backend
}

const HOURS_PER_DAY = 24

const DAYS_PER_WEEK = 7

const DAYS_PER_BIWEEK = 14

const DAYS_PER_MONTH = 30

const DAYS_PER_QUARTER = 3 * 30

const DAYS_PER_BIANNUM = 6 * 30

const DAYS_PER_ANNUM = 12 * 30

func SectionToTimestamp(section string, startTime time.Time) time.Time {
	switch section {
	case "daily":
		return startTime.Add(HOURS_PER_DAY * time.Hour)
	case "weekly":
		return startTime.Add(DAYS_PER_WEEK * HOURS_PER_DAY * time.Hour)
	case "bi-weekly":
		return startTime.Add(DAYS_PER_BIWEEK * HOURS_PER_DAY * time.Hour)
	case "monthly":
		return startTime.Add(DAYS_PER_MONTH * HOURS_PER_DAY * time.Hour)
	case "quarterly":
		return startTime.Add(DAYS_PER_QUARTER * HOURS_PER_DAY * time.Hour)
	case "bi-annually":
		return startTime.Add(DAYS_PER_BIANNUM * HOURS_PER_DAY * time.Hour)
	default:
		// We will default to `montly` to be safe
		return startTime.Add(time.Duration(DAYS_PER_MONTH*HOURS_PER_DAY) * time.Hour)
	}
}
