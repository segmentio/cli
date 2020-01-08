// Package human provides types that support parsing and formatting
// human-friendly representations of values in various units.
package human

import (
	"fmt"
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

func parseUnit(s string) (head, unit string) {
	i := strings.LastIndexFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	if i < 0 {
		head = s
		return
	}

	head = strings.TrimRightFunc(s[:i+1], unicode.IsSpace)
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
