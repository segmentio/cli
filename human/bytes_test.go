package human

import (
	"encoding/json"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestBytesParse(t *testing.T) {
	for _, test := range []struct {
		in  string
		out Bytes
	}{
		{in: "0", out: 0},

		{in: "2B", out: 2},
		{in: "2K", out: 2 * KB},
		{in: "2M", out: 2 * MB},
		{in: "2G", out: 2 * GB},
		{in: "2T", out: 2 * TB},
		{in: "2P", out: 2 * PB},

		{in: "2", out: 2},
		{in: "2Ki", out: 2 * KiB},
		{in: "2Mi", out: 2 * MiB},
		{in: "2Gi", out: 2 * GiB},
		{in: "2Ti", out: 2 * TiB},
		{in: "2Pi", out: 2 * PiB},

		{in: "1.234K", out: 1234},
		{in: "1.234M", out: 1234 * KB},

		{in: "1.5Ki", out: 1*KiB + 512},
		{in: "1.5Mi", out: 1*MiB + 512*KiB},
	} {
		t.Run(test.in, func(t *testing.T) {
			b, err := ParseBytes(test.in)
			if err != nil {
				t.Fatal(err)
			}
			if b != test.out {
				t.Error("parsed bytes mismatch:", b, "!=", test.out)
			}
		})
	}
}

func TestBytesFormat(t *testing.T) {
	for _, test := range []struct {
		in  Bytes
		out string
	}{
		{out: "0", in: 0},
		{out: "2", in: 2},

		{out: "1.953Ki", in: 2 * KB},
		{out: "1.907Mi", in: 2 * MB},
		{out: "1.863Gi", in: 2 * GB},
		{out: "1.819Ti", in: 2 * TB},
		{out: "1.776Pi", in: 2 * PB},

		{out: "2Ki", in: 2 * KiB},
		{out: "2Mi", in: 2 * MiB},
		{out: "2Gi", in: 2 * GiB},
		{out: "2Ti", in: 2 * TiB},
		{out: "2Pi", in: 2 * PiB},

		{out: "1.205Ki", in: 1234},
		{out: "1.177Mi", in: 1234 * KB},

		{out: "1.5Ki", in: 1*KiB + 512},
		{out: "1.5Mi", in: 1*MiB + 512*KiB},
	} {
		t.Run(test.out, func(t *testing.T) {
			if s := test.in.String(); s != test.out {
				t.Error("formatted bytes mismatch:", s, "!=", test.out)
			}
		})
	}
}

func TestBytesJSON(t *testing.T) {
	testBytesEncoding(t, 1*KiB, json.Marshal, json.Unmarshal)
}

func TestBytesYAML(t *testing.T) {
	testBytesEncoding(t, 1*KiB, yaml.Marshal, yaml.Unmarshal)
}

func testBytesEncoding(t *testing.T, x Bytes, marshal func(interface{}) ([]byte, error), unmarshal func([]byte, interface{}) error) {
	b, err := marshal(x)
	if err != nil {
		t.Fatal("marshal error:", err)
	}

	v := Bytes(0)
	if err := unmarshal(b, &v); err != nil {
		t.Error("unmarshal error:", err)
	} else if v != x {
		t.Error("value mismatch:", v, "!=", x)
	}
}
