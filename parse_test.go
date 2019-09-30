package cli

import (
	"reflect"
	"testing"
)

func TestParseCommandLine(t *testing.T) {
	parser := parser{
		options: map[string]option{
			"-A":     {boolean: false},
			"--bool": {boolean: true},
		},
	}

	args := []string{
		"-A=1", "-A", "2", "--bool", "a", "b", "c", "--", "command", "line",
	}

	options, values, command, err := parser.parseCommandLine(args)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(options, map[string][]string{
		"-A":     {"1", "2"},
		"--bool": {"true"},
	}) {
		t.Error("options mismatch:", options)
	}

	if !reflect.DeepEqual(values, []string{"a", "b", "c"}) {
		t.Error("values mismatch:", values)
	}

	if !reflect.DeepEqual(command, []string{"command", "line"}) {
		t.Error("command mismatch:", command)
	}
}
