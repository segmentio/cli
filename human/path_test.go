package human

import (
	"os/user"
	"path/filepath"
	"testing"
)

func TestPath(t *testing.T) {
	separator := string([]byte{filepath.Separator})
	user, _ := user.Current()
	home := user.HomeDir

	tests := []struct {
		in  string
		out Path
	}{
		{in: ".", out: "."},
		{in: separator, out: Path(separator)},
		{in: filepath.Join(".", "hello", "world"), out: Path(filepath.Join(".", "hello", "world"))},
		{in: filepath.Join("~", "hello", "world"), out: Path(filepath.Join(home, "hello", "world"))},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			path := Path("")

			if err := path.UnmarshalText([]byte(test.in)); err != nil {
				t.Error(err)
			} else if path != test.out {
				t.Errorf("path mismatch: %q != %q", path, test.out)
			}
		})
	}
}
