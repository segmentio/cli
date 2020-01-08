package human

import (
	"encoding"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	yaml "gopkg.in/yaml.v3"
)

type Time time.Time

func ParseTime(s string) (Time, error) {
	return ParseTimeAt(s, time.Now())
}

func ParseTimeAt(s string, now time.Time) (Time, error) {
	if strings.HasSuffix(s, " ago") {
		s = strings.TrimLeftFunc(s[:len(s)-4], unicode.IsSpace)
		d, err := ParseDurationUntil(s, now)
		if err != nil {
			return Time{}, fmt.Errorf("malformed time representation: %q", s)
		}
		return Time(now.Add(-time.Duration(d))), nil
	}

	if strings.HasSuffix(s, " later") {
		s = strings.TrimRightFunc(s[:len(s)-6], unicode.IsSpace)
		d, err := ParseDurationUntil(s, now)
		if err != nil {
			return Time{}, fmt.Errorf("malformed time representation: %q", s)
		}
		return Time(now.Add(time.Duration(d))), nil
	}

	for _, format := range []string{
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
	} {
		t, err := time.Parse(format, s)
		if err == nil {
			return Time(t), nil
		}
	}

	return Time{}, fmt.Errorf("unsupported time representation: %q", s)
}

func (t Time) IsZero() bool {
	return time.Time(t).IsZero()
}

func (t Time) String() string {
	return t.Text(time.Now())
}

func (t Time) Text(now time.Time) string {
	if t.IsZero() {
		return "(none)"
	}
	d := now.Sub(time.Time(t))
	s := ""
	if d >= 0 {
		s = " ago"
	} else {
		s = " later"
		d = -d
	}
	return Duration(d).Text(now) + s
}

func (t Time) MarshalJSON() ([]byte, error) {
	return time.Time(t).MarshalJSON()
}

func (t *Time) UnmarshalJSON(b []byte) error {
	return ((*time.Time)(t)).UnmarshalJSON(b)
}

func (t Time) MarshalYAML() (interface{}, error) {
	return time.Time(t).Format(time.RFC3339Nano), nil
}

func (t *Time) UnmarshalYAML(y *yaml.Node) error {
	var s string
	if err := y.Decode(&s); err != nil {
		return err
	}
	p, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return err
	}
	*t = Time(p)
	return nil
}

func (t Time) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *Time) UnmarshalText(b []byte) error {
	p, err := ParseTime(string(b))
	if err != nil {
		return err
	}
	*t = p
	return nil
}

var (
	_ fmt.Stringer = Time{}

	_ json.Marshaler   = Time{}
	_ json.Unmarshaler = (*Time)(nil)

	_ yaml.IsZeroer    = Time{}
	_ yaml.Marshaler   = Time{}
	_ yaml.Unmarshaler = (*Time)(nil)

	_ encoding.TextMarshaler   = Time{}
	_ encoding.TextUnmarshaler = (*Time)(nil)
)
