package human

import (
	"encoding/json"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestDurationParse(t *testing.T) {
	for _, test := range []struct {
		in  string
		out Duration
	}{
		{in: "0", out: 0},

		{in: "1ns", out: Nanosecond},
		{in: "1µs", out: Microsecond},
		{in: "1ms", out: Millisecond},
		{in: "1s", out: Second},
		{in: "1m", out: Minute},
		{in: "1h", out: Hour},

		{in: "1d", out: 24 * Hour},
		{in: "2d", out: 48 * Hour},
		{in: "1w", out: 7 * 24 * Hour},
		{in: "2w", out: 14 * 24 * Hour},

		{in: "1 nanosecond", out: Nanosecond},
		{in: "1 microsecond", out: Microsecond},
		{in: "1 millisecond", out: Millisecond},
		{in: "1 second", out: Second},
		{in: "1 minute", out: Minute},
		{in: "1 hour", out: Hour},

		{in: "1 day", out: 24 * Hour},
		{in: "2 days", out: 48 * Hour},
		{in: "1 week", out: 7 * 24 * Hour},
		{in: "2 weeks", out: 14 * 24 * Hour},

		{in: "1m30s", out: 1*Minute + 30*Second},
		{in: "1.5m", out: 1*Minute + 30*Second},
	} {
		t.Run(test.in, func(t *testing.T) {
			d, err := ParseDuration(test.in)
			if err != nil {
				t.Fatal(err)
			}
			if d != test.out {
				t.Error("parsed duration mismatch:", d, "!=", test.out)
			}
		})
	}
}

func TestDurationFormat(t *testing.T) {
	for _, test := range []struct {
		in  Duration
		out string
	}{
		{out: "0s", in: 0},

		{out: "1ns", in: Nanosecond},
		{out: "1µs", in: Microsecond},
		{out: "1ms", in: Millisecond},
		{out: "1s", in: Second},
		{out: "1m", in: Minute},
		{out: "1h", in: Hour},

		{out: "1d", in: 24 * Hour},
		{out: "2d", in: 48 * Hour},
		{out: "1w", in: 7 * 24 * Hour},
		{out: "2w", in: 14 * 24 * Hour},
		{out: "1mo", in: 33 * 24 * Hour},
		{out: "2mo", in: 66 * 24 * Hour},
		{out: "1y", in: 400 * 24 * Hour},
		{out: "2y", in: 800 * 24 * Hour},
	} {
		t.Run(test.out, func(t *testing.T) {
			if s := test.in.String(); s != test.out {
				t.Error("duration string mismatch:", s, "!=", test.out)
			}
		})
	}
}

func TestDurationJSON(t *testing.T) {
	testDurationEncoding(t, (2 * Hour), json.Marshal, json.Unmarshal)
}

func TestDurationYAML(t *testing.T) {
	testDurationEncoding(t, (2 * Hour), yaml.Marshal, yaml.Unmarshal)
}

func testDurationEncoding(t *testing.T, x Duration, marshal func(interface{}) ([]byte, error), unmarshal func([]byte, interface{}) error) {
	b, err := marshal(x)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	v := Duration(0)
	if err := unmarshal(b, &v); err != nil {
		t.Error("unmarshal error:", err)
	} else if v != x {
		t.Error("value mismatch:", v, "!=", x)
	}
}
