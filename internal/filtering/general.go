package filter

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var skippablePrefixes = []string{"#", "mailto"}

var skippableExtensions = []string{
	"acc",
	"bmp",
	"doc",
	"docx",
	"epub",
	"gif",
	"jpeg",
	"jpg",
	"mov",
	"mp3",
	"ogg",
	"png",
	"psd",
	"raw",
	"stl",
	"svg",
	"tif",
	"tiff",
	"txt",
	"webp",
	"xls",
	"xlsx",
}

const IsTooShortMin = 5

const ExceedsLengthMax = 200

const TooManyRepeatsLen = 8

const TooManyRepeatsCount = 50

func exceedsLength(length int) func(*url.URL) error {
	return func(u *url.URL) error {
		if len(u.String()) > length {
			return fmt.Errorf("exceeds length [%d]: %s", length, u.String())
		}

		return nil
	}
}

func hasSlashHTTP(u *url.URL) error {
	m, _ := regexp.MatchString(`/http`, u.Path)
	if m {
		return fmt.Errorf("http in middle of url: %s", u.Path)
	}

	return nil
}

func insecureGov(u *url.URL) error {
	m, _ := regexp.MatchString(`^http:`, u.String())
	if m {
		return fmt.Errorf("insecure URL: %s", u.String())
	}

	return nil
}

func isTooShort(length int) func(*url.URL) error {
	return func(u *url.URL) error {
		if len(u.String()) < length {
			return fmt.Errorf("too short [%d]: %s", length, u.String())
		}

		return nil
	}
}

func hasSkippablePrefixRelative(u *url.URL) error {
	for _, sp := range skippablePrefixes {
		if strings.HasPrefix(u.String(), sp) {
			return fmt.Errorf("skippable prefix [%s]: %s", sp, u.Path)
		}
	}

	return nil
}

func hasSkippableExtension(u *url.URL) error {
	for _, ext := range skippableExtensions {
		if strings.HasSuffix(u.Path, ext) {
			return fmt.Errorf("skippable extension [%s]: %s", ext, u.Path)
		}
	}

	return nil
}

func hasTooManyRepeats(repeatLength int, threshold int) func(*url.URL) error {
	return func(u *url.URL) error {
		s := u.String()

		end := len(s) - repeatLength
		chunks := make(map[string]bool)
		repeats := make(map[string]int)

		for ndx := 0; ndx < end; ndx++ {
			piece := s[ndx : ndx+repeatLength]
			if _, ok := chunks[piece]; ok {
				repeats[piece] = repeats[piece] + 1
			} else {
				chunks[piece] = true
				repeats[piece] = 0
			}
		}

		total := 0
		for _, v := range repeats {
			total += v
		}

		if total >= threshold {
			return fmt.Errorf("too many repeats [%d over %d]: %s", total, threshold, u.String())
		}

		return nil
	}
}

func endsWithWrongSlash(u *url.URL) error {
	// log.Println("URL LOOKS LIKE", u.String())
	for _, pat := range []string{`\$`, `%5C$`} {
		m, _ := regexp.MatchString(pat, u.String())
		if m {
			return fmt.Errorf("ends with backslash: %s", u.String())
		}
	}

	return nil
}

var all = ".*"

func GeneralRules() []Rule {
	rules := make([]Rule, 0)

	rules = append(rules, Rule{
		Match:  all,
		Msg:    "insecure gov",
		Reject: insecureGov,
	})

	rules = append(rules, Rule{
		Match:  all,
		Msg:    "max isTooShort 5",
		Reject: isTooShort(IsTooShortMin),
	})

	rules = append(rules, Rule{
		Match:  all,
		Msg:    "exceedsLength 200",
		Reject: exceedsLength(ExceedsLengthMax),
	})

	rules = append(rules, Rule{
		Match:  all,
		Msg:    "endsWithWrongSlash",
		Reject: endsWithWrongSlash,
	})

	rules = append(rules, Rule{
		Match:  all,
		Msg:    "hasSlashHttp",
		Reject: hasSlashHTTP,
	})

	rules = append(rules, Rule{
		Match:  all,
		Msg:    "insecureGov",
		Reject: insecureGov,
	})

	rules = append(rules, Rule{
		Match:  all,
		Msg:    "hasSkippablePrefixRelative",
		Reject: hasSkippablePrefixRelative,
	})

	rules = append(rules, Rule{
		Match:  all,
		Msg:    "hasSkippableExtension",
		Reject: hasSkippableExtension,
	})

	rules = append(rules, Rule{
		Match:  all,
		Msg:    "hasTooManyRepeats",
		Reject: hasTooManyRepeats(TooManyRepeatsLen, TooManyRepeatsCount),
	})

	return rules
}
