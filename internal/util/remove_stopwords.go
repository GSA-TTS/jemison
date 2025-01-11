package util

import (
	_ "embed"
	"regexp"
	"slices"
	"strings"
)

//go:embed stopwords.txt
var stopwords string

var each_stopword []string

var ws_re = regexp.MustCompile(`\s+`)

var punc_re = regexp.MustCompile(`[-_\.!\?,]`)

func removeStopwords(content string) string {
	content = ws_re.ReplaceAllString(content, " ")
	each := strings.Split(content, " ")
	new_content := make([]string, 0)

	for _, e := range each {
		e = punc_re.ReplaceAllString(e, " ")

		if !slices.Contains(each_stopword, e) {
			new_content = append(new_content, e)
		}
	}

	return ws_re.ReplaceAllString(strings.Join(new_content, " "), " ")
}

func RemoveStopwords(content string) string {
	if len(each_stopword) > 0 {
		return removeStopwords(content)
	} else {
		each_stopword = strings.Split(stopwords, "\n")

		return removeStopwords(content)
	}
}
