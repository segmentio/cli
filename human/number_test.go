package human

import (
	"encoding/json"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestNumberParse(t *testing.T) {
	for _, test := range []struct {
		in  string
		out Number
	}{
		{in: "0", out: 0},
		{in: "1234", out: 1234},
		{in: "1,234", out: 1234},
		{in: "1,234.567", out: 1234.567},
	} {
		t.Run(test.in, func(t *testing.T) {
			n, err := ParseNumber(test.in)
			if err != nil {
				t.Fatal(err)
			}
			if n != test.out {
				t.Error("parsed number mismatch:", n, "!=", test.out)
			}
		})
	}
}

func TestNumberFormat(t *testing.T) {
	for _, test := range []struct {
		in  Number
		out string
	}{
		{in: 0, out: "0"},
		{in: 1234, out: "1,234"},
		{in: 1234.567, out: "1,234.567"},
		{in: 123456.789, out: "123,456.789"},
		{in: 1234567.89, out: "1,234,567.89"},
	} {
		t.Run(test.out, func(t *testing.T) {
			if s := test.in.String(); s != test.out {
				t.Error("formatted number mismatch:", s, "!=", test.out)
			}
		})
	}
}

func TestNumberJSON(t *testing.T) {
	testNumberEncoding(t, Number(1.234), json.Marshal, json.Unmarshal)
}

func TestNumberYAML(t *testing.T) {
	testNumberEncoding(t, Number(1.234), yaml.Marshal, yaml.Unmarshal)
}

func testNumberEncoding(t *testing.T, x Number, marshal func(interface{}) ([]byte, error), unmarshal func([]byte, interface{}) error) {
	b, err := marshal(x)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	v := Number(0)
	if err := unmarshal(b, &v); err != nil {
		t.Error("unmarshal error:", err)
	} else if v != x {
		t.Error("value mismatch:", v, "!=", x)
	}
}
