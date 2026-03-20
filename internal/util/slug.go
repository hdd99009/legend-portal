package util

import (
	"regexp"
	"strings"
)

var nonSlugPattern = regexp.MustCompile(`[^a-z0-9]+`)

func Slugify(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	s = strings.ReplaceAll(s, "_", "-")
	s = nonSlugPattern.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "post"
	}
	return s
}
