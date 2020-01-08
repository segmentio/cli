package human

import (
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	yaml "gopkg.in/yaml.v3"
)

type Bytes uint64

const (
	B Bytes = 1

	KB Bytes = 1000 * B
	MB Bytes = 1000 * KB
	GB Bytes = 1000 * MB
	TB Bytes = 1000 * GB
	PB Bytes = 1000 * TB

	KiB Bytes = 1024 * B
	MiB Bytes = 1024 * KiB
	GiB Bytes = 1024 * MiB
	TiB Bytes = 1024 * GiB
	PiB Bytes = 1024 * TiB
)

func ParseBytes(s string) (Bytes, error) {
	f, err := ParseBytesFloat64(s)
	if err != nil {
		return 0, err
	}
	if f < 0 {
		return 0, fmt.Errorf("invalid negative byte count: %q", s)
	}
	return Bytes(math.Floor(f)), err
}

func ParseBytesFloat64(s string) (float64, error) {
	value, unit := parseUnit(s)

	scale := Bytes(0)
	switch {
	case match(unit, "B"), unit == "":
		scale = B
	case match(unit, "KB"):
		scale = KB
	case match(unit, "MB"):
		scale = MB
	case match(unit, "GB"):
		scale = GB
	case match(unit, "TB"):
		scale = TB
	case match(unit, "PB"):
		scale = PB
	case match(unit, "KiB"):
		scale = KiB
	case match(unit, "MiB"):
		scale = MiB
	case match(unit, "GiB"):
		scale = GiB
	case match(unit, "TiB"):
		scale = TiB
	case match(unit, "PiB"):
		scale = PiB
	default:
		return 0, fmt.Errorf("malformed bytes representation: %q", s)
	}

	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("malformed bytes representations: %q: %w", s, err)
	}
	return f * float64(scale), nil
}

func (b Bytes) String() string {
	var scale Bytes
	var unit string

	switch {
	case b >= PiB:
		scale, unit = PiB, "Pi"
	case b >= TiB:
		scale, unit = TiB, "Ti"
	case b >= GiB:
		scale, unit = GiB, "Gi"
	case b >= MiB:
		scale, unit = MiB, "Mi"
	case b >= KiB:
		scale, unit = KiB, "Ki"
	default:
		scale, unit = B, ""
	}

	s := fmt.Sprintf("%.3f", float64(b)/float64(scale))
	s = suffix('0').trim(s)
	s = suffix('.').trim(s)
	return s + unit
}

func (b Bytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint64(b))
}

func (b *Bytes) UnmarshalJSON(j []byte) error {
	return json.Unmarshal(j, (*uint64)(b))
}

func (b Bytes) MarshalYAML() (interface{}, error) {
	return b.String(), nil
}

func (b *Bytes) UnmarshalYAML(y *yaml.Node) error {
	var s string
	if err := y.Decode(&s); err != nil {
		return err
	}
	p, err := ParseBytes(s)
	if err != nil {
		return err
	}
	*b = p
	return nil
}

func (b Bytes) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

func (b *Bytes) UnmarshalText(t []byte) error {
	p, err := ParseBytes(string(t))
	if err != nil {
		return err
	}
	*b = p
	return nil
}

var (
	_ fmt.Stringer = Bytes(0)

	_ json.Marshaler   = Bytes(0)
	_ json.Unmarshaler = (*Bytes)(nil)

	_ yaml.Marshaler   = Bytes(0)
	_ yaml.Unmarshaler = (*Bytes)(nil)

	_ encoding.TextMarshaler   = Bytes(0)
	_ encoding.TextUnmarshaler = (*Bytes)(nil)
)
