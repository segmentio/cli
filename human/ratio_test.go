package human

import (
	"encoding/json"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestRatioParse(t *testing.T) {
	for _, test := range []struct {
		in  string
		out Ratio
	}{
		{in: "0", out: 0},
		{in: "0%", out: 0},
		{in: "0.0%", out: 0},
		{in: "12.34%", out: 0.1234},
		{in: "100%", out: 1},
		{in: "200%", out: 2},
	} {
		t.Run(test.in, func(t *testing.T) {
			n, err := ParseRatio(test.in)
			if err != nil {
				t.Fatal(err)
			}
			if n != test.out {
				t.Error("parsed ratio mismatch:", n, "!=", test.out)
			}
		})
	}
}

func TestRatioFormat(t *testing.T) {
	for _, test := range []struct {
		in  Ratio
		out string
	}{
		{in: 0, out: "0%"},
		{in: 0.1234, out: "12.34%"},
		{in: 1, out: "100%"},
		{in: 2, out: "200%"},
	} {
		t.Run(test.out, func(t *testing.T) {
			if s := test.in.String(); s != test.out {
				t.Error("formatted ratio mismatch:", s, "!=", test.out)
			}
		})
	}
}

func TestRatioJSON(t *testing.T) {
	testRatioEncoding(t, Ratio(0.234), json.Marshal, json.Unmarshal)
}

func TestRatioYAML(t *testing.T) {
	testRatioEncoding(t, Ratio(0.234), yaml.Marshal, yaml.Unmarshal)
}

func testRatioEncoding(t *testing.T, x Ratio, marshal func(interface{}) ([]byte, error), unmarshal func([]byte, interface{}) error) {
	b, err := marshal(x)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	v := Ratio(0)
	if err := unmarshal(b, &v); err != nil {
		t.Error("unmarshal error:", err)
	} else if v != x {
		t.Error("value mismatch:", v, "!=", x)
	}
}
