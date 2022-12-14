package cli

import (
	"reflect"
	"testing"
)

type testStructType struct {
	_        struct{} `help:"Greets someone from a galaxy far, far away"`
	Name     string   `flag:"-n,--name" help:"Someone's name" default:"Luke"`
	Surname  string   `flag:"-s, --surname, --last-name" help:"Someone's surname" default:"Skywalker"`
	Planet   string   `flag:"-p,--planet" help:"Someone's home planet" env:"-" default:"-"`
	darkside bool     `flag:"--dark,--dark-side" help:"True if friend of the Sith"`
}

func TestForEachStructField(t *testing.T) {
	testStruct := testStructType{}
	structType := reflect.TypeOf(testStruct)

	var foundName bool
	var foundSurname bool
	var foundPlanet bool
	forEachStructField(structType, nil, func(sf structField) {
		if sf.typ.Kind() != reflect.String {
			t.Errorf("Type of field expected to be string, got %s", sf.typ)
		}
		switch sf.help {
		case "Someone's name":
			if len(sf.flags) != 2 || sf.flags[0] != "-n" || sf.flags[1] != "--name" {
				t.Errorf("Incorrect flags for Name field: %v", sf.flags)
			}
			if len(sf.envvars) != 1 || sf.envvars[0] != "NAME" {
				t.Errorf("Incorrect envvars for Name field: %v", sf.envvars)
			}
			if sf.defval != "Luke" {
				t.Errorf("Incorrect default value for Name field: %s", sf.defval)
			}
			foundName = true
		case "Someone's surname":
			if len(sf.flags) != 3 || sf.flags[0] != "-s" || sf.flags[1] != "--surname" || sf.flags[2] != "--last-name" {
				t.Errorf("Incorrect flags for Surname field: %s", sf.flags)
			}
			if len(sf.envvars) != 2 || sf.envvars[0] != "SURNAME" || sf.envvars[1] != "LAST_NAME" {
				t.Errorf("Incorrect envvars for Surname field: %v", sf.envvars)
			}
			if sf.defval != "Skywalker" {
				t.Errorf("Incorrect default value for Surname field: %s", sf.defval)
			}
			foundSurname = true
		case "Someone's home planet":
			if len(sf.flags) != 2 || sf.flags[0] != "-p" || sf.flags[1] != "--planet" {
				t.Errorf("Incorrect flags for Planet field: %v", sf.flags)
			}
			if len(sf.envvars) != 0 {
				t.Errorf("Incorrect envvars for Planet field: %v", sf.envvars)
			}
			if sf.defval != "-" {
				t.Errorf("Incorrect default value for Planet field: %s", sf.defval)
			}
			foundPlanet = true
		default:
			// _ should be skipped because it is command help
			// darkside should be skipped because it is not exported
			t.Fatalf("Found unexpected field, help text: %s", sf.help)
		}
	})

	if !foundName {
		t.Error("Failed to locate Name field")
	}
	if !foundSurname {
		t.Error("Failed to locate Surname field")
	}
	if !foundPlanet {
		t.Error("Failed to locate Planet field")
	}
}
