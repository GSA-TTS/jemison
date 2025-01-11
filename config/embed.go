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

func ReadJSONConfig(jsonFilename string) string {
	jsonBytes, err := ConfigFs.ReadFile(jsonFilename)
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	return string(jsonBytes)
}

func GetYamlFileReader(yamlFilename string) *bytes.Reader {
	yamlBytes, err := ConfigFs.ReadFile(yamlFilename)
	if err != nil {
		zap.L().Fatal(err.Error())
	}

	return bytes.NewReader(yamlBytes)
}

func GetListOfHosts(allowedHosts string) []string {
	zap.L().Debug("reading in hosts", zap.String("allowed_hosts", allowedHosts))

	cfg := ReadJSONConfig("allowed_hosts.yaml")

	// The variable `allowed_hosts` will be the key into the doc that has
	// a list of pairs. Each pair is a range of values, which tells us how
	// to filter the FQDN/D64 values.
	ranges := gjson.Get(cfg, allowedHosts).Array()
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
	cfg := ReadJSONConfig(schedule)
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

const HoursPerDay = 24

const DaysPerWeek = 7

const DaysPerBiWeek = 14

const DaysPerMonth = 30

const DaysPerQuarter = 3 * 30

const DaysPerBiAnnum = 6 * 30

const DaysPerAnnum = 12 * 30

func SectionToTimestamp(section string, startTime time.Time) time.Time {
	switch section {
	case "daily":
		return startTime.Add(HoursPerDay * time.Hour)
	case "weekly":
		return startTime.Add(DaysPerWeek * HoursPerDay * time.Hour)
	case "bi-weekly":
		return startTime.Add(DaysPerBiWeek * HoursPerDay * time.Hour)
	case "monthly":
		return startTime.Add(DaysPerMonth * HoursPerDay * time.Hour)
	case "quarterly":
		return startTime.Add(DaysPerQuarter * HoursPerDay * time.Hour)
	case "bi-annually":
		return startTime.Add(DaysPerBiAnnum * HoursPerDay * time.Hour)
	default:
		// We will default to `montly` to be safe
		return startTime.Add(time.Duration(DaysPerMonth*HoursPerDay) * time.Hour)
	}
}
