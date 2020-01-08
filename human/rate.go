package human

import (
	"encoding"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	yaml "gopkg.in/yaml.v3"
)

// Rate represents a count devided by a unit of time.
//
// The type supports parsing and formatting values like:
//
//	200/s
//	1 / minute
//	0.5/week
//	...
//
// Rate values are always stored in their per-second form in Go programs,
// and properly converted during parsing and formatting.
type Rate float64

const (
	PerNanosecond  Rate = 1 / Rate(Nanosecond)
	PerMicrosecond Rate = 1 / Rate(Microsecond)
	PerMillisecond Rate = 1 / Rate(Millisecond)
	PerSecond      Rate = 1 / Rate(Second)
	PerMinute      Rate = 1 / Rate(Minute)
	PerHour        Rate = 1 / Rate(Hour)
	PerDay         Rate = 1 / Rate(Day)
	PerWeek        Rate = 1 / Rate(Week)
)

func ParseRate(s string) (Rate, error) {
	var text string
	var unit string
	var rate Rate

	if i := strings.IndexByte(s, '/'); i < 0 {
		text = s
	} else {
		text = strings.TrimLeftFunc(s[:i], unicode.IsSpace)
		unit = strings.TrimRightFunc(s[i+1:], unicode.IsSpace)
	}

	c, err := ParseCount(text)
	if err != nil {
		return 0, fmt.Errorf("malformed rate representation: %q", s)
	}

	switch {
	case match(unit, "week"):
		rate = PerWeek
	case match(unit, "day"):
		rate = PerDay
	case match(unit, "hour"):
		rate = PerHour
	case match(unit, "minute"):
		rate = PerMinute
	case match(unit, "second"), unit == "":
		rate = PerSecond
	case match(unit, "millisecond"), unit == "ms", unit == "µs":
		rate = PerMillisecond
	case match(unit, "microsecond"), unit == "us":
		rate = PerMicrosecond
	case match(unit, "nanosecond"), unit == "ns":
		rate = PerNanosecond
	default:
		return 0, fmt.Errorf("malformed unit representation: %q", s)
	}

	return Rate(c) * (rate / PerSecond), nil
}

func (r Rate) String() string {
	return r.Text(Second)
}

func (r Rate) Text(d Duration) string {
	var unit string

	switch {
	case d >= Week:
		unit = "/w"
	case d >= Day:
		unit = "/d"
	case d >= Hour:
		unit = "/h"
	case d >= Minute:
		unit = "/m"
	case d >= Second:
		unit = "/s"
	case d >= Millisecond:
		unit = "/ms"
	case d >= Microsecond:
		unit = "/µs"
	default:
		unit = "/ns"
	}

	r /= Rate(d) * PerSecond
	return Count(r).String() + unit
}

func (r Rate) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(r))
}

func (r *Rate) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, (*float64)(r))
}

func (r Rate) MarshalYAML() (interface{}, error) {
	return r.String(), nil
}

func (r *Rate) UnmarshalYAML(y *yaml.Node) error {
	var s string
	if err := y.Decode(&s); err != nil {
		return err
	}
	p, err := ParseRate(s)
	if err != nil {
		return err
	}
	*r = p
	return nil
}

func (r Rate) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *Rate) UnmarshalText(b []byte) error {
	p, err := ParseRate(string(b))
	if err != nil {
		return err
	}
	*r = p
	return nil
}

var (
	_ fmt.Stringer = Rate(0)

	_ json.Marshaler   = Rate(0)
	_ json.Unmarshaler = (*Rate)(nil)

	_ yaml.Marshaler   = Rate(0)
	_ yaml.Unmarshaler = (*Rate)(nil)

	_ encoding.TextMarshaler   = Rate(0)
	_ encoding.TextUnmarshaler = (*Rate)(nil)
)
