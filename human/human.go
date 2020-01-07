// Package human provides types that support parsing and formatting
// human-friendly representations of values in various units.
package human

import (
	"strconv"
	"strings"
	"unicode"
)

func parseNextToken(s string) (string, string) {
	for i, r := range s {
		if unicode.IsSpace(r) {
			return s[:i], strings.TrimLeftFunc(s[i:], unicode.IsSpace)
		}
	}
	return s, ""
}

func parseInt(s string) (int, string, error) {
	s, r := parseNextToken(s)
	i, err := strconv.Atoi(s)
	return i, r, err
}

func match(s, pattern string) bool {
	return strings.HasPrefix(pattern, s)
}
