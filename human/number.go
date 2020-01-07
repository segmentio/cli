package human

import (
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type Number float64

func ParseNumber(s string) (Number, error) {
	r := strings.ReplaceAll(s, ",", "")
	f, err := strconv.ParseFloat(r, 64)
	if err != nil {
		return 0, fmt.Errorf("malformed number: %s: %w", s, err)
	}
	return Number(f), nil
}

func (n Number) String() string {
	if n == 0 {
		return "0"
	}

	if n < 0 {
		return "-" + (-n).String()
	}

	if n <= 1e-3 || n >= 1e12 {
		return strconv.FormatFloat(float64(n), 'g', -1, 64)
	}

	i, d := math.Modf(float64(n))
	parts := make([]string, 0, 4)

	for u := uint64(i); u > 0; u /= 1000 {
		parts = append(parts, strconv.FormatUint(u%1000, 10))
	}

	for i, j := 0, len(parts)-1; i < j; {
		parts[i], parts[j] = parts[j], parts[i]
		i++
		j--
	}

	r := strings.Join(parts, ",")

	if d != 0 {
		r += "."
		r += trimZeroSuffix(strconv.FormatUint(uint64(math.Round(d*1000)), 10))
	}

	return r
}

func trimZeroSuffix(s string) string {
	for len(s) != 0 && s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}
	return s
}

func (n Number) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(n))
}

func (n *Number) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, (*float64)(n))
}

func (n Number) MarshalYAML() (interface{}, error) {
	return float64(n), nil
}

func (n *Number) UnmarshalYAML(y *yaml.Node) error {
	return y.Decode((*float64)(n))
}

func (n Number) MarshalText() ([]byte, error) {
	return []byte(n.String()), nil
}

func (n *Number) UnmarshalText(b []byte) error {
	p, err := ParseNumber(string(b))
	if err != nil {
		return err
	}
	*n = p
	return nil
}

var (
	_ fmt.Stringer = Number(0)

	_ json.Marshaler   = Number(0)
	_ json.Unmarshaler = (*Number)(nil)

	_ yaml.Marshaler   = Number(0)
	_ yaml.Unmarshaler = (*Number)(nil)

	_ encoding.TextMarshaler   = Number(0)
	_ encoding.TextUnmarshaler = (*Number)(nil)
)
