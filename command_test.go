package cli

import "testing"

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
