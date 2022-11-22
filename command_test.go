package cli

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/segmentio/cli/human"
)

func TestSimilarEnough(t *testing.T) {
	tests := []struct {
		input, cmd string
		want       bool
	}{
		{"a", "d", false},           // 1
		{"ab", "cd", false},         // 2
		{"abc", "bbc", true},        // 1
		{"abcd", "bbcd", true},      // 1
		{"abcd", "bbcc", false},     // 2
		{"abcde", "bbcdd", false},   // 2
		{"abcdef", "bbcddf", false}, // 2
		{"abcdef", "abcddf", true},  // 1
		{"bbcdfg", "abcdefg", true}, // 2
	}
	for _, tt := range tests {
		lvn := levenshtein(tt.input, tt.cmd)
		got := similarEnough(tt.input, tt.cmd, lvn)
		if got != tt.want {
			t.Errorf("similarEnough(%q, %q, %d): got %t, want %t", tt.input,
				tt.cmd, lvn, got, tt.want)
		}
	}
}

// The individual struct decoders don't have access to the entire command, but
// it should be assigned by the parent *CommandFunc after the error is caught.
func TestStructDecoderFail(t *testing.T) {
	var b bytes.Buffer
	Err = &b
	defer func() {
		Err = os.Stderr
	}()
	type config struct {
		Duration human.Duration `flag:"--duration"`
	}
	cmd := Command(func(cfg config) {
		fmt.Println(cfg.Duration)
	})
	Call(cmd, "--duration=10")
	want := `
Usage:
  [options]

Options:
      --duration duration
  -h, --help               Show this help message

Error:
  decoding "--duration": please include a unit ('weeks', 'h', 'm') in addition to the value (10.000000)


`
	if b.String() != want {
		t.Errorf("Struct error: got %q, want %q", b.String(), want)
	}
}
