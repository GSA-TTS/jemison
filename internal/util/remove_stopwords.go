package util

import (
	_ "embed"
	"regexp"
	"slices"
	"strings"
)

//go:embed stopwords.txt
var stopwords string

var eachStopword []string

var wsRe = regexp.MustCompile(`\s+`)

var puncRe = regexp.MustCompile(`[-_\.!\?,]`)

func removeStopwords(content string) string {
	content = wsRe.ReplaceAllString(content, " ")
	each := strings.Split(content, " ")
	newContent := make([]string, 0)

	for _, e := range each {
		e = puncRe.ReplaceAllString(e, " ")

		if !slices.Contains(eachStopword, e) {
			newContent = append(newContent, e)
		}
	}

	return wsRe.ReplaceAllString(strings.Join(newContent, " "), " ")
}

func RemoveStopwords(content string) string {
	if len(eachStopword) > 0 {
		return removeStopwords(content)
	}

	eachStopword = strings.Split(stopwords, "\n")

	return removeStopwords(content)
}
