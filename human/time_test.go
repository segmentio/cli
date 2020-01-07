package human

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTimeParse(t *testing.T) {
	now := time.Now()

	for _, test := range []struct {
		in  string
		out Duration
	}{
		{in: "0s ago", out: Duration(0)},

		{in: "1ns ago", out: Duration(-time.Nanosecond)},
		{in: "1µs ago", out: Duration(-time.Microsecond)},
		{in: "1ms ago", out: Duration(-time.Millisecond)},
		{in: "1s ago", out: Duration(-time.Second)},
		{in: "1m ago", out: Duration(-time.Minute)},
		{in: "1h ago", out: Duration(-time.Hour)},

		{in: "1 nanosecond ago", out: Duration(-time.Nanosecond)},
		{in: "1 microsecond ago", out: Duration(-time.Microsecond)},
		{in: "1 millisecond ago", out: Duration(-time.Millisecond)},
		{in: "1 second ago", out: Duration(-time.Second)},
		{in: "1 minute ago", out: Duration(-time.Minute)},
		{in: "1 hour ago", out: Duration(-time.Hour)},

		{in: "1 day ago", out: Duration(-24 * time.Hour)},
		{in: "2 days ago", out: Duration(-48 * time.Hour)},
		{in: "1 week ago", out: Duration(-7 * 24 * time.Hour)},
		{in: "2 weeks ago", out: Duration(-14 * 24 * time.Hour)},

		{in: "0s later", out: Duration(0)},

		{in: "1ns later", out: Duration(time.Nanosecond)},
		{in: "1µs later", out: Duration(time.Microsecond)},
		{in: "1ms later", out: Duration(time.Millisecond)},
		{in: "1s later", out: Duration(time.Second)},
		{in: "1m later", out: Duration(time.Minute)},
		{in: "1h later", out: Duration(time.Hour)},

		{in: "1 nanosecond later", out: Duration(time.Nanosecond)},
		{in: "1 microsecond later", out: Duration(time.Microsecond)},
		{in: "1 millisecond later", out: Duration(time.Millisecond)},
		{in: "1 second later", out: Duration(time.Second)},
		{in: "1 minute later", out: Duration(time.Minute)},
		{in: "1 hour later", out: Duration(time.Hour)},

		{in: "1 day later", out: Duration(24 * time.Hour)},
		{in: "2 days later", out: Duration(48 * time.Hour)},
		{in: "1 week later", out: Duration(7 * 24 * time.Hour)},
		{in: "2 weeks later", out: Duration(14 * 24 * time.Hour)},
	} {
		t.Run(test.in, func(t *testing.T) {
			p, err := ParseTimeAt(test.in, now)
			if err != nil {
				t.Fatal(err)
			}
			if d := Duration(time.Time(p).Sub(now)); d != test.out {
				t.Error("parsed time delta mismatch:", d, "!=", test.out)
			}

			u := Time(now.Add(time.Duration(test.out)))
			v := Time{}

			b, err := json.Marshal(u)
			if err != nil {
				t.Fatal("json marshal error:", err)
			}
			if err := json.Unmarshal(b, &v); err != nil {
				t.Error("json unmarshal error:", err)
			} else if !time.Time(v).Equal(time.Time(u)) {
				t.Error("json value mismatch:", v, "!=", u)
			}
		})
	}
}

func TestTimeFormat(t *testing.T) {
	now := time.Now()

	for _, test := range []struct {
		in  Duration
		out string
	}{
		{out: "0s ago", in: Duration(0)},

		{out: "1ns ago", in: Duration(-time.Nanosecond)},
		{out: "1µs ago", in: Duration(-time.Microsecond)},
		{out: "1ms ago", in: Duration(-time.Millisecond)},
		{out: "1s ago", in: Duration(-time.Second)},
		{out: "1m ago", in: Duration(-time.Minute)},
		{out: "1h ago", in: Duration(-time.Hour)},

		{out: "1 day ago", in: Duration(-24 * time.Hour)},
		{out: "2 days ago", in: Duration(-48 * time.Hour)},
		{out: "1 week ago", in: Duration(-7 * 24 * time.Hour)},
		{out: "2 weeks ago", in: Duration(-14 * 24 * time.Hour)},
		{out: "1 month ago", in: Duration(-33 * 24 * time.Hour)},
		{out: "2 months ago", in: Duration(-66 * 24 * time.Hour)},
		{out: "1 year ago", in: Duration(-400 * 24 * time.Hour)},
		{out: "2 years ago", in: Duration(-800 * 24 * time.Hour)},

		{out: "1ns later", in: Duration(time.Nanosecond)},
		{out: "1µs later", in: Duration(time.Microsecond)},
		{out: "1ms later", in: Duration(time.Millisecond)},
		{out: "1s later", in: Duration(time.Second)},
		{out: "1m later", in: Duration(time.Minute)},
		{out: "1h later", in: Duration(time.Hour)},

		{out: "1 day later", in: Duration(24 * time.Hour)},
		{out: "2 days later", in: Duration(48 * time.Hour)},
		{out: "1 week later", in: Duration(7 * 24 * time.Hour)},
		{out: "2 weeks later", in: Duration(14 * 24 * time.Hour)},
		{out: "1 month later", in: Duration(33 * 24 * time.Hour)},
		{out: "2 months later", in: Duration(66 * 24 * time.Hour)},
		{out: "1 year later", in: Duration(400 * 24 * time.Hour)},
		{out: "2 years later", in: Duration(800 * 24 * time.Hour)},
	} {
		t.Run(test.out, func(t *testing.T) {
			if s := Time(now.Add(time.Duration(test.in))).StringAt(now); s != test.out {
				t.Error("time string mismatch:", s, "!=", test.out)
			}
		})
	}
}
