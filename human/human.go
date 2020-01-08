// Package human provides types that support parsing and formatting
// human-friendly representations of values in various units.
package human

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func isDot(r rune) bool {
	return r == '.'
}

func isExp(r rune) bool {
	return r == 'e' || r == 'E'
}

func isSign(r rune) bool {
	return r == '-' || r == '+'
}

func isNumberPrefix(r rune) bool {
	return isSign(r) || unicode.IsDigit(r)
}

func hasPrefixFunc(s string, f func(rune) bool) bool {
	for _, r := range s {
		return f(r)
	}
	return false
}

func countPrefixFunc(s string, f func(rune) bool) int {
	var i int
	var r rune

	for i, r = range s {
		if !f(r) {
			break
		}
	}

	return i
}

func skipSpaces(s string) string {
	return strings.TrimLeftFunc(s, unicode.IsSpace)
}

func trimSpaces(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

func parseNextNumber(s string) (string, string) {
	i := 0

	// integer part
	i += countPrefixFunc(s[i:], isSign) // - or +
	i += countPrefixFunc(s[i:], unicode.IsDigit)

	// decimal part
	if hasPrefixFunc(s[i:], isDot) {
		i++ // .
		i += countPrefixFunc(s[i:], unicode.IsDigit)
	}

	// exponent part
	if hasPrefixFunc(s[i:], isExp) {
		i++                                 // e or E
		i += countPrefixFunc(s[i:], isSign) // - or +
		i += countPrefixFunc(s[i:], unicode.IsDigit)
	}

	return s[:i], skipSpaces(s[i:])
}

func parseNextToken(s string) (string, string) {
	if hasPrefixFunc(s, isNumberPrefix) {
		return parseNextNumber(s)
	}

	for i, r := range s {
		if unicode.IsSpace(r) {
			return s[:i], skipSpaces(s[i:])
		}
	}

	return s, ""
}

func parseInt(s string) (int, string, error) {
	s, r := parseNextToken(s)
	i, err := strconv.Atoi(s)
	return i, r, err
}

func parseUnit(s string) (head, unit string) {
	i := strings.LastIndexFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	if i < 0 {
		head = s
		return
	}

	head = trimSpaces(s[:i+1])
	unit = s[i+1:]
	return
}

func match(s, pattern string) bool {
	return len(s) <= len(pattern) && strings.EqualFold(s, pattern[:len(s)])
}

type prefix byte

func (c prefix) trim(s string) string {
	for len(s) > 0 && s[0] == byte(c) {
		s = s[1:]
	}
	return s
}

type suffix byte

func (c suffix) trim(s string) string {
	for len(s) > 0 && s[len(s)-1] == byte(c) {
		s = s[:len(s)-1]
	}
	return s
}

func (c suffix) match(s string) bool {
	return len(s) > 0 && s[len(s)-1] == byte(c)
}

func fabs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func ftoa(value, scale float64) string {
	var format string

	if value == 0 {
		return "0"
	}

	if value < 0 {
		return "-" + ftoa(-value, scale)
	}

	switch {
	case (value / scale) >= 100:
		format = "%.0f"
	case (value / scale) >= 10:
		format = "%.1f"
	case scale > 1:
		format = "%.2f"
	default:
		format = "%.3f"
	}

	s := fmt.Sprintf(format, value/scale)
	if strings.Contains(s, ".") {
		s = suffix('0').trim(s)
		s = suffix('.').trim(s)
	}
	return s
}
