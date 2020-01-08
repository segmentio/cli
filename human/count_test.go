package human

import (
	"encoding/json"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestCountParse(t *testing.T) {
	for _, test := range []struct {
		in  string
		out Count
	}{
		{in: "0", out: 0},
		{in: "1234", out: 1234},
		{in: "10.2K", out: 10200},
	} {
		t.Run(test.in, func(t *testing.T) {
			c, err := ParseCount(test.in)
			if err != nil {
				t.Fatal(err)
			}
			if c != test.out {
				t.Error("parsed count mismatch:", c, "!=", test.out)
			}
		})
	}
}

func TestCountFormat(t *testing.T) {
	for _, test := range []struct {
		in  Count
		out string
	}{
		{in: 0, out: "0"},
		{in: 1234, out: "1234"},
		{in: 10234, out: "10.2K"},
	} {
		t.Run(test.out, func(t *testing.T) {
			if s := test.in.String(); s != test.out {
				t.Error("formatted count mismatch:", s, "!=", test.out)
			}
		})
	}
}

func TestCountJSON(t *testing.T) {
	testCountEncoding(t, Count(1.234), json.Marshal, json.Unmarshal)
}

func TestCountYAML(t *testing.T) {
	testCountEncoding(t, Count(1.234), yaml.Marshal, yaml.Unmarshal)
}

func testCountEncoding(t *testing.T, x Count, marshal func(interface{}) ([]byte, error), unmarshal func([]byte, interface{}) error) {
	b, err := marshal(x)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	v := Count(0)
	if err := unmarshal(b, &v); err != nil {
		t.Error("unmarshal error:", err)
	} else if v != x {
		t.Error("value mismatch:", v, "!=", x)
	}
}
