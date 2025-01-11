package filter

import (
	"fmt"
	"net/url"
	"regexp"
)

// We compare against the host, so leave off the scheme.
var nasa = `.*nasa.gov`

func hasRightHere(u *url.URL) error {
	match, _ := regexp.MatchString("right.*?here", u.String())
	if match {
		return fmt.Errorf("repeating `right here`: %s", u.String())
	}

	return nil
}

func NasaRules() []Rule {
	rules := make([]Rule, 0)

	rules = append(rules, Rule{
		Match:  nasa,
		Msg:    "recursive self reference (right%20here)",
		Reject: hasRightHere,
	})

	return rules
}
