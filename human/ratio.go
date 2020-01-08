package human

import (
	"encoding"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// Ratio represents percentage-like values.
//
// The type supports parsing and formatting values like:
//
//	0.1
//	25%
//	0.5 %
//	...
//
// Ratio values are stored as floating pointer numbers between 0 and 1 (assuming
// they stay within the 0-100% bounds), and formatted as percentages.
type Ratio float64

func ParseRatio(s string) (Ratio, error) {
	k := 1.0
	p := suffix('%')

	if p.match(s) {
		k = 100.0
		s = trimSpaces(s[:len(s)-1])
	}

	f, err := strconv.ParseFloat(s, 64)
	return Ratio(f / k), err
}

func (r Ratio) String() string {
	return r.Text(2)
}

func (r Ratio) Text(precision int) string {
	s := strconv.FormatFloat(100*float64(r), 'f', precision, 64)
	if strings.Contains(s, ".") {
		s = suffix('0').trim(s)
		s = suffix('.').trim(s)
	}
	return s + "%"
}

func (r Ratio) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(r))
}

func (r *Ratio) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, (*float64)(r))
}

func (r Ratio) MarshalYAML() (interface{}, error) {
	return r.Text(-1), nil
}

func (r *Ratio) UnmarshalYAML(y *yaml.Node) error {
	var s string
	if err := y.Decode(&s); err != nil {
		return err
	}
	p, err := ParseRatio(s)
	if err != nil {
		return err
	}
	*r = Ratio(p)
	return nil
}

func (r Ratio) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *Ratio) UnmarshalText(b []byte) error {
	p, err := ParseRatio(string(b))
	if err != nil {
		return err
	}
	*r = p
	return nil
}

var (
	_ fmt.Stringer = Ratio(0)

	_ json.Marshaler   = Ratio(0)
	_ json.Unmarshaler = (*Ratio)(nil)

	_ yaml.Marshaler   = Ratio(0)
	_ yaml.Unmarshaler = (*Ratio)(nil)

	_ encoding.TextMarshaler   = Ratio(0)
	_ encoding.TextUnmarshaler = (*Ratio)(nil)
)
