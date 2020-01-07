package human

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDurationParse(t *testing.T) {
	for _, test := range []struct {
		in  string
		out Duration
	}{
		{in: "0", out: Duration(0)},

		{in: "1ns", out: Duration(time.Nanosecond)},
		{in: "1µs", out: Duration(time.Microsecond)},
		{in: "1ms", out: Duration(time.Millisecond)},
		{in: "1s", out: Duration(time.Second)},
		{in: "1m", out: Duration(time.Minute)},
		{in: "1h", out: Duration(time.Hour)},

		{in: "1 nanosecond", out: Duration(time.Nanosecond)},
		{in: "1 microsecond", out: Duration(time.Microsecond)},
		{in: "1 millisecond", out: Duration(time.Millisecond)},
		{in: "1 second", out: Duration(time.Second)},
		{in: "1 minute", out: Duration(time.Minute)},
		{in: "1 hour", out: Duration(time.Hour)},

		{in: "1 day", out: Duration(24 * time.Hour)},
		{in: "2 days", out: Duration(48 * time.Hour)},
		{in: "1 week", out: Duration(7 * 24 * time.Hour)},
		{in: "2 weeks", out: Duration(14 * 24 * time.Hour)},
	} {
		t.Run(test.in, func(t *testing.T) {
			d, err := ParseDuration(test.in)
			if err != nil {
				t.Fatal(err)
			}
			if d != test.out {
				t.Error("parsed duration mismatch:", d, "!=", test.out)
			}

			u := test.out
			v := Duration(0)

			b, err := json.Marshal(u)
			if err != nil {
				t.Fatal("json marshal error:", err)
			}
			if err := json.Unmarshal(b, &v); err != nil {
				t.Error("json unmarshal error:", err)
			} else if v != u {
				t.Error("json value mismatch:", v, "!=", u)
			}
		})
	}
}

func TestDurationFormat(t *testing.T) {
	for _, test := range []struct {
		in  Duration
		out string
	}{
		{out: "0s", in: Duration(0)},

		{out: "1ns", in: Duration(time.Nanosecond)},
		{out: "1µs", in: Duration(time.Microsecond)},
		{out: "1ms", in: Duration(time.Millisecond)},
		{out: "1s", in: Duration(time.Second)},
		{out: "1m", in: Duration(time.Minute)},
		{out: "1h", in: Duration(time.Hour)},

		{out: "1 day", in: Duration(24 * time.Hour)},
		{out: "2 days", in: Duration(48 * time.Hour)},
		{out: "1 week", in: Duration(7 * 24 * time.Hour)},
		{out: "2 weeks", in: Duration(14 * 24 * time.Hour)},
		{out: "1 month", in: Duration(33 * 24 * time.Hour)},
		{out: "2 months", in: Duration(66 * 24 * time.Hour)},
		{out: "1 year", in: Duration(400 * 24 * time.Hour)},
		{out: "2 years", in: Duration(800 * 24 * time.Hour)},
	} {
		t.Run(test.out, func(t *testing.T) {
			if s := test.in.String(); s != test.out {
				t.Error("duration string mismatch:", s, "!=", test.out)
			}
		})
	}
}
