package human

import (
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"time"

	yaml "gopkg.in/yaml.v3"
)

const (
	Nanosecond  Duration = 1
	Microsecond Duration = 1000 * Nanosecond
	Millisecond Duration = 1000 * Microsecond
	Second      Duration = 1000 * Millisecond
	Minute      Duration = 60 * Second
	Hour        Duration = 60 * Minute
	Day         Duration = 24 * Hour
	Week        Duration = 7 * Day
)

// Duration is based on time.Duration, but supports parsing and formatting
// more human-friendly representations.
//
// Here are examples of supported values:
//
//	1d
//	4 weeks
//	...
//
// The current implementation does not support decimal values, however,
// contributions are welcome to add this feature.
//
// Time being what it is, months and years are hard to represent because their
// durations vary in unpredictable ways. This is why the package only exposes
// constants up to a 1 week duration. For the sake of accuracy, years and months
// are always represented relative to a given date. Technically, leap seconds
// can cause any unit above the second to be variable, but in order to remain
// mentaly sane, we chose to ignore this detail in the implementation of this
// package.
type Duration time.Duration

func ParseDuration(s string) (Duration, error) {
	return ParseDurationUntil(s, time.Now())
}

func ParseDurationUntil(s string, now time.Time) (Duration, error) {
	var d Duration
	var input = s

	for len(s) != 0 {
		p, err := time.ParseDuration(s)
		if err == nil {
			d += Duration(p)
			break
		}

		// TODO: we should probably support floating point numbers here to
		// represent values like "1.5 years".
		n, r, err := parseInt(s)
		if err != nil {
			return 0, fmt.Errorf("malformed duration: %s: %w", input, err)
		}
		s = r

		v, r, err := parseDuration(s, n, now)
		if err != nil {
			return 0, fmt.Errorf("malformed duration: %s: %w", input, err)
		}
		s = r

		d += v
	}

	return d, nil
}

func parseDuration(s string, n int, now time.Time) (Duration, string, error) {
	s, r := parseNextToken(s)
	switch {
	case match(s, "weeks"):
		return Duration(n) * Week, r, nil
	case match(s, "days"):
		return Duration(n) * Day, r, nil
	case match(s, "hours"):
		return Duration(n) * Hour, r, nil
	case match(s, "minutes"):
		return Duration(n) * Minute, r, nil
	case match(s, "seconds"):
		return Duration(n) * Second, r, nil
	case match(s, "milliseconds"), s == "ms":
		return Duration(n) * Millisecond, r, nil
	case match(s, "microseconds"), s == "us", s == "µs":
		return Duration(n) * Microsecond, r, nil
	case match(s, "nanoseconds"), s == "ns":
		return Duration(n) * Nanosecond, r, nil
	case match(s, "months"):
		return Duration(now.AddDate(0, n, 0).Sub(now)), r, nil
	case match(s, "years"):
		return Duration(now.AddDate(n, 0, 0).Sub(now)), r, nil
	default:
		return 0, "", fmt.Errorf("unkonwn time unit %q", s)
	}
}

func (d Duration) String() string {
	return d.Text(time.Now())
}

func (d Duration) Text(until time.Time) string {
	if d == 0 {
		return "0s"
	}

	if d < 0 {
		if d == Duration(math.MinInt64) {
			return "-272y"
		}
		return "-" + (-d).Text(until)
	}

	if d < 31*Day {
		switch {
		case d < Microsecond:
			return fmt.Sprintf("%dns", d)
		case d < Millisecond:
			return fmt.Sprintf("%dµs", d/Microsecond)
		case d < Second:
			return fmt.Sprintf("%dms", d/Millisecond)
		case d < Minute:
			return fmt.Sprintf("%ds", d/Second)
		case d < Hour:
			return fmt.Sprintf("%dm", d/Minute)
		case d < Day:
			return fmt.Sprintf("%dh", d/Hour)
		case d < Week:
			return fmt.Sprintf("%dd", d/Day)
		default:
			return fmt.Sprintf("%dw", d/Week)
		}
	}

	if y := d.Years(until); y != 0 {
		return fmt.Sprintf("%dy", y)
	}

	return fmt.Sprintf("%dmo", d.Months(until))
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d))
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, (*time.Duration)(d))
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

func (d *Duration) UnmarshalYAML(y *yaml.Node) error {
	var s string
	if err := y.Decode(&s); err != nil {
		return err
	}
	p, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(p)
	return nil
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Duration) UnmarshalText(b []byte) error {
	p, err := ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = p
	return nil
}

func (d Duration) Nanoseconds() int { return int(d) }

func (d Duration) Microseconds() int { return int(d) / int(Microsecond) }

func (d Duration) Milliseconds() int { return int(d) / int(Millisecond) }

func (d Duration) Seconds() int { return int(d) / int(Second) }

func (d Duration) Minutes() int { return int(d) / int(Minute) }

func (d Duration) Hours() int { return int(d) / int(Hour) }

func (d Duration) Days() int { return int(d) / int(Day) }

func (d Duration) Weeks() int { return int(d) / int(Week) }

func (d Duration) Months(until time.Time) int {
	if d < 0 {
		return -((-d).Months(until.Add(-time.Duration(d))))
	}

	cursor := until.Add(-time.Duration(d + 1))
	months := 0

	for cursor.Before(until) {
		cursor = cursor.AddDate(0, 1, 0)
		months++
	}

	return months - 1
}

func (d Duration) Years(until time.Time) int {
	if d < 0 {
		return -((-d).Years(until.Add(-time.Duration(d))))
	}

	cursor := until.Add(-time.Duration(d + 1))
	years := 0

	for cursor.Before(until) {
		cursor = cursor.AddDate(1, 0, 0)
		years++
	}

	return years - 1
}

var (
	_ fmt.Stringer = Duration(0)

	_ json.Marshaler   = Duration(0)
	_ json.Unmarshaler = (*Duration)(nil)

	_ yaml.Marshaler   = Duration(0)
	_ yaml.Unmarshaler = (*Duration)(nil)

	_ encoding.TextMarshaler   = Duration(0)
	_ encoding.TextUnmarshaler = (*Duration)(nil)
)
