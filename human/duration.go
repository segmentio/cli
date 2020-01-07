package human

import (
	"encoding"
	"encoding/json"
	"fmt"
	"time"
)

const (
	nanosecond  = Duration(time.Nanosecond)
	microsecond = Duration(time.Microsecond)
	millisecond = Duration(time.Millisecond)
	second      = Duration(time.Second)
	minute      = Duration(time.Minute)
	hour        = Duration(time.Hour)
	day         = 24 * hour
	week        = 7 * day
)

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
		return Duration(n) * week, r, nil
	case match(s, "days"):
		return Duration(n) * day, r, nil
	case match(s, "hours"):
		return Duration(n) * hour, r, nil
	case match(s, "minutes"):
		return Duration(n) * minute, r, nil
	case match(s, "seconds"):
		return Duration(n) * second, r, nil
	case match(s, "milliseconds") || s == "ms":
		return Duration(n) * millisecond, r, nil
	case match(s, "microseconds") || s == "us" || s == "µs":
		return Duration(n) * microsecond, r, nil
	case match(s, "nanoseconds") || s == "ns":
		return Duration(n) * nanosecond, r, nil
	case match(s, "months"):
		return Duration(now.AddDate(0, n, 0).Sub(now)), r, nil
	case match(s, "years"):
		return Duration(now.AddDate(n, 0, 0).Sub(now)), r, nil
	default:
		return 0, "", fmt.Errorf("unkonwn time unit %q", s)
	}
}

func (d Duration) String() string {
	return d.StringUntil(time.Now())
}

func (d Duration) StringUntil(until time.Time) string {
	if d == 0 {
		return "0s"
	}

	if d < 31*day {
		switch {
		case d < microsecond:
			return fmt.Sprintf("%dns", d)
		case d < millisecond:
			return fmt.Sprintf("%dµs", d/microsecond)
		case d < second:
			return fmt.Sprintf("%dms", d/millisecond)
		case d < minute:
			return fmt.Sprintf("%ds", d/second)
		case d < hour:
			return fmt.Sprintf("%dm", d/minute)
		case d < day:
			return fmt.Sprintf("%dh", d/hour)
		case d >= day && d < 2*day:
			return "1 day"
		case d < week:
			return fmt.Sprintf("%d days", d/day)
		case d >= week && d < 2*week:
			return "1 week"
		default:
			return fmt.Sprintf("%d weeks", d/week)
		}
	}

	switch years := d.Years(until); {
	case years == 1:
		return "1 year"
	case years > 1:
		return fmt.Sprintf("%d years", years)
	}

	switch months := d.Months(until); {
	case months == 1:
		return "1 month"
	default:
		return fmt.Sprintf("%d months", months)
	}
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d))
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, (*time.Duration)(d))
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

func (d Duration) Microseconds() int { return int(d) / int(microsecond) }

func (d Duration) Milliseconds() int { return int(d) / int(millisecond) }

func (d Duration) Seconds() int { return int(d) / int(second) }

func (d Duration) Minutes() int { return int(d) / int(minute) }

func (d Duration) Hours() int { return int(d) / int(hour) }

func (d Duration) Days() int { return int(d) / int(day) }

func (d Duration) Weeks() int { return int(d) / int(week) }

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

	_ encoding.TextMarshaler   = Duration(0)
	_ encoding.TextUnmarshaler = (*Duration)(nil)
)
