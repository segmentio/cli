package cli

import (
	"fmt"
	"strings"
)

type option struct {
	boolean bool
}

type parser struct {
	aliases map[string]string
	options map[string]option
}

func makeParser() parser {
	return parser{
		aliases: map[string]string{"-h": "--help"},
		options: map[string]option{"--help": {boolean: true}},
	}
}

func (p parser) parseCommandLine(args []string) (options map[string][]string, values, command []string, err error) {
	options = make(map[string][]string)

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if isCommandSeparator(arg) { // command after "--"
			command = append([]string{}, args[i+1:]...)
			break
		}

		if !isOption(arg) { // positional argument
			values = append(values, arg)
			continue
		}

		name, value, hasValue := splitNameValue(arg)
		// If the argument is an alias, overwrite with the main option name to
		// ensure that all values given for that option are combined.
		alias, ok := p.aliases[name]
		if ok {
			name = alias
		}

		option, ok := p.options[name]
		if !ok {
			err = &Usage{Err: fmt.Errorf("unrecognized option: %q", arg)}
			return
		}

		if option.boolean {
			if hasValue {
				switch value {
				case "true", "false":
				default:
					err = &Usage{Err: fmt.Errorf("unexpected boolean value: %q", value)}
					return
				}
			} else {
				value, hasValue = "true", true
			}
		}

		if hasValue { // option=value
			options[name] = append(options[name], value)
			continue
		}

		if i++; i == len(args) || isOption(args[i]) {
			err = &Usage{Err: fmt.Errorf("missing option value: %q", arg)}
			return
		}

		options[name] = append(options[name], args[i])
	}

	return
}

func isOption(s string) bool {
	return len(s) > 1 && s[0] == '-'
}

func isCommandSeparator(s string) bool {
	return s == "--"
}

func splitNameValue(s string) (string, string, bool) {
	i := strings.IndexByte(s, '=')
	if i < 0 {
		return s, "", false
	}
	return s[:i], s[i+1:], true
}

func lookupEnv(name string, env []string) (string, bool) {
	for _, e := range env {
		if k, v, _ := splitNameValue(e); k == name {
			return v, true
		}
	}
	return "", false
}
