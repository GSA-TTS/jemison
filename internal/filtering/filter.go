package filter

import (
	"fmt"
	"net/url"
	"regexp"

	"go.uber.org/zap"
)

type Rule struct {
	Match  string
	Msg    string
	Reject func(*url.URL) error
}

func GetRules() []Rule {
	rules := make([]Rule, 0)
	rules = append(rules, GeneralRules()...)
	rules = append(rules, NasaRules()...)

	return rules
}

func IsRejectRuleset(u *url.URL, rules []Rule) error {
	failed := false
	failedMsg := ""

	var e error

	for _, r := range rules {
		apply, err := regexp.MatchString(r.Match, u.Host)
		if err != nil {
			return fmt.Errorf("should not get here: %s", err.Error())
		}

		if apply {
			// log.Println("applying", r.Msg)
			err := r.Reject(u)
			if err != nil {
				zap.L().Debug("reject based on rule",
					zap.String("msg", r.Msg))

				failed = true
				failedMsg = r.Msg
				e = err

				break
			}
		}
	}

	if failed {
		return fmt.Errorf("%s: %s", failedMsg, e.Error())
	}

	return nil
}

func IsReject(u *url.URL) error {
	rules := GetRules()

	return IsRejectRuleset(u, rules)
}
