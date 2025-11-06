package utils

import (
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`\s+`)

func RemoveWhitespace(s string) string {
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}
