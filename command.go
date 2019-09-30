package cli

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"text/tabwriter"
	"unicode/utf8"
)

// Command constructs a Function which delegates to the Go function passed as
// argument.
//
// The argument is an interface{} because functions with multiple types are
// accepted.
//
// The function may receive no arguments, indicating that the program delegating
// to the command is expeceted to be invoked with no arguments, for example:
//
//	cmd := cli.Command(func() {
//		...
//	})
//
// If the function accepts arguments, the first argument must always be a struct
// type, which describes the set of options that are accepted by the command:
//
//	// Struct tags are used to declare the flags accepted by the command.
//	type config struct {
//		Path    string `flag:"--path"       help:"Path to a text file" default:"file.txt"`
//		Verbose bool   `flag:"-v,--verbose" help:"Enable verbose mode"`
//	}
//
//	cmd := cli.Command(func(config config) {
//		...
//	})
//
// Three keys are recognized in the struct tags: "flag", "help", and "default".
//
// The "flag" struct tag is a comma-separated list of command line flags that
// map to the field. This tag is required.
//
// The "help" struct tag is a human-redable message describing what the field is
// used for.
//
// The "default" struct tag provides the default value of the field when the
// argument was missing from the call to the command. Any flag which has no
// default value and isn't a boolean or a slice type must be passed when calling
// the command, otherwise a usage error is returned. The special default value
// "-" can be used to indicate that the option is not required and should assume
// its zero-value when omitted.
//
// If the struct contains a field named `_`, the command will look for a "help"
// struct tag to define its own help message. Note that the type of the field
// is irrelevant, but it is common practice to use an empty struct.
//
// The command always injects and handles the -h and --help flags, which can be
// used to request that call to the command return a help error to describe the
// configuration options of the command.
//
// Every flag starting with a "--" may also be configured via an environment
// variable. The environment variable is matched by converting the flag name to
// a snakecase and uppercase format. Flags that should not be matched to
// environment variables must specify a struct tag env:"-" to disable the
// feature.
//
// Each extra argument to the function is interpreted as a positional argument
// and decoded as such, for example:
//
//	// This command expects two integers as positional arguments.
//	cmd := cli.Command(func(config config, x, y int) {
//		...
//	})
//
// The last positional argument may be a slice, which consumes as many values as
// remained on the command invocation.
//
// An extra variadic string parameter may be accepted by the function, which
// receives any extra arguments found after a "--" separator. This mechanism is
// often used by programs that spawn other programs to define the limits between
// the arguments of the first program, and the second command.
//
// If the command is called with an invalid set of arguments, it returns a
// non-zero code and a usage error which describes the issue.
func Command(fn interface{}) Function { return &CommandFunc{Func: fn} }

// CommandFunc is an implementation of the Function interface which calls out to
// a function when invoked.
type CommandFunc struct {
	// A short help message describing what the command does.
	Help string

	// A full description of the command.
	Desc string

	// The function that the command calls out to when invoked.
	//
	// See Command for details about the accepted signatures.
	Func interface{}

	function reflect.Value
	parser   parser
	options  structDecoder
	values   []decodeFunc
	variadic bool
	help     string
}

func (cmd *CommandFunc) configure() {
	if cmd.function.IsValid() {
		return // already configured
	}

	t := reflect.TypeOf(cmd.Func)
	v := reflect.ValueOf(cmd.Func)

	if t.Kind() != reflect.Func {
		panic("cli.Command: expected a function as argument but got " + t.String())
	}

	cmd.function = v
	cmd.variadic = t.IsVariadic()

	if n := t.NumIn(); n == 0 {
		cmd.parser, cmd.options, cmd.help = makeStructDecoder(emptyType)
	} else {
		if f := t.In(0); f.Kind() == reflect.Struct {
			cmd.parser, cmd.options, cmd.help = makeStructDecoder(f)
		} else {
			panic("cli.Command: expected a struct as first argument but got " + f.String())
		}

		if cmd.variadic {
			n--
		}

		for i := 1; i < n; i++ {
			p := t.In(i)

			if p.Kind() == reflect.Slice {
				cmd.values = append(cmd.values, makeSliceDecoder(p))
				break
			}

			cmd.values = append(cmd.values, makeValueDecoder(p))
		}
	}

	switch t.NumOut() {
	case 0:
	case 1:
		if r0 := t.Out(0); r0 != errorType {
			panic("cli.Command: expected a function returning (error) but got (" + r0.String() + ")")
		}
	case 2:
		if r0, r1 := t.Out(0), t.Out(1); r0 != intType || r1 != errorType {
			panic("cli.Command: expected a function returing (int, error) but got (" + r0.String() + ", " + r1.String() + ")")
		}
	default:
		panic("cli.Command: the function returns too many values")
	}

	if cmd.help == "" {
		cmd.help = cmd.Help
	}
}

// Call satisfies the Function interface.
//
// See Command for the full documentation of how the Call method behaves.
func (cmd *CommandFunc) Call(args, env []string) (int, error) {
	cmd.configure()

	options, values, command, err := cmd.parser.parseCommandLine(args)
	if err != nil {
		return 1, err
	}

	if wantHelp(options) {
		return 0, &Help{Cmd: cmd}
	}

	for name, field := range cmd.options {
		if _, ok := options[name]; !ok && len(field.envvars) != 0 {
			for _, e := range field.envvars {
				if v, ok := lookupEnv(e, env); ok {
					options[name] = []string{v}
					break
				}
			}
		}
	}

	for name, field := range cmd.options {
		if _, ok := options[name]; !ok && field.defval != "" && field.defval != "-" {
			options[name] = []string{field.defval}
		}
	}

	for name, field := range cmd.options {
		if _, ok := options[name]; !ok && field.defval == "" && !field.boolean && !field.slice {
			return 1, &Usage{Cmd: cmd, Err: fmt.Errorf("missing required option: %q", name)}
		}
	}

	var params []reflect.Value

	if t := cmd.function.Type(); t.NumIn() > 0 {
		// Configuration options are decoded into the first function parameter.
		v := reflect.New(t.In(0)).Elem()
		if err := cmd.options.decode(v, options); err != nil {
			return 1, err
		}
		params = append(params, v)

		// Positional arguments are decoded into each following function
		// parameter, until a slice type is encountered which receives all
		// the remaining values.
		n := t.NumIn()

		if cmd.variadic {
			n--
		}

		for i := 1; i < n; i++ {
			p := t.In(i)
			v := reflect.New(p).Elem()

			if p.Kind() == reflect.Slice {
				if err := cmd.values[i-1](v, values); err != nil {
					return 1, err
				}
				params = append(params, v)
				values = nil
				break
			}

			if len(values) == 0 {
				return 1, &Usage{
					Cmd: cmd,
					Err: fmt.Errorf("expected %d positional arguments but only %d were given", len(cmd.values), i-1),
				}
			}

			if err := cmd.values[i-1](v, values[:1]); err != nil {
				return 1, err
			}
			params = append(params, v)
			values = values[1:]
		}
	}

	if len(values) != 0 {
		return 1, &Usage{
			Cmd: cmd,
			Err: fmt.Errorf("too many positional arguments: %q", values),
		}
	}

	if cmd.variadic && len(command) == 0 {
		return 1, &Usage{
			Cmd: cmd,
			Err: fmt.Errorf("missing command after \"--\" separator"),
		}
	}

	if !cmd.variadic && len(command) != 0 {
		return 1, &Usage{
			Cmd: cmd,
			Err: fmt.Errorf("unsupported command after \"--\" separator"),
		}
	}

	var r []reflect.Value
	if cmd.variadic {
		r = cmd.function.CallSlice(append(params, reflect.ValueOf(command)))
	} else {
		r = cmd.function.Call(params)
	}

	var ret int
	switch len(r) {
	case 0:
	case 1:
		if err, _ = r[0].Interface().(error); err != nil {
			ret = 1
		}
	default:
		ret, _ = r[0].Interface().(int)
		err, _ = r[1].Interface().(error)
	}

	switch e := err.(type) {
	case *Help:
		e.Cmd = cmd
	case *Usage:
		e.Cmd = cmd
	}

	return ret, err
}

// Format statisfies the fmt.Formatter interface, its recognizes the following
// verbs:
//
//	%s	outputs the usage information of the command
//	%v	outputs the full description of the command
//	%x	outputs the help message of the command
//
func (cmd *CommandFunc) Format(w fmt.State, v rune) {
	switch v {
	case 's': // usage
		io.WriteString(w, "[options]")

		t := cmd.function.Type()
		n := t.NumIn()
		if cmd.variadic {
			n--
		}

		for i := 1; i < n; i++ {
			p := t.In(i)
			fmt.Fprintf(w, " [%s]", typeNameOf(p))

			if p.Kind() == reflect.Slice {
				break
			}
		}

		if cmd.variadic {
			io.WriteString(w, " -- [command]")
		}

	case 'v': // description
		if cmd.Desc != "" {
			for _, line := range strings.Split(cmd.Desc, "\n") {
				fmt.Fprintf(w, "  %s\n", line)
			}
			io.WriteString(w, "\n")
		}

		io.WriteString(w, "Options:\n")

		tw := newTabWriter(w)
		defer tw.Flush()

		// Compute the length of all short flags in order to align the positions
		// of short and long flags on different columns.
		shortLen := 0

		for _, field := range cmd.options {
			n := 0
			for _, f := range field.flags {
				if isShortFlag(f) {
					n += utf8.RuneCountInString(f) + 2
				}
			}
			if n > shortLen {
				shortLen = n
			}
		}

		b := &bytes.Buffer{}
		b.Grow(128)

		for _, fieldName := range sortedMapKeys(reflect.ValueOf(cmd.options)) {
			b.Reset()
			b.WriteString("  ") // indent
			field := cmd.options[fieldName.String()]

			// This counter is used to track how many short and long flags have
			// been written.
			//
			// Short flags are printed first, then long flags. Empty columes are
			// written between short and long flags to align fields.
			n := 0

			for i, f := range field.flags {
				if isShortFlag(f) {
					n += writeFlag(b, f, i, len(field.flags))
				}
			}

			for n < shortLen {
				b.WriteByte(' ')
				n++
			}

			for i, f := range field.flags {
				if isLongFlag(f) {
					writeFlag(b, f, i, len(field.flags))
				}
			}

			if field.argtyp != "" {
				b.WriteString(" ")
				b.WriteString(field.argtyp)
			}

			b.WriteString("\t")

			if field.help != "" {
				b.WriteString("  ")
				b.WriteString(field.help)
			}

			if field.defval != "" && field.defval != "-" {
				fmt.Fprintf(b, " (default: %s)", field.defval)
			}

			b.WriteString("\n")
			tw.Write(b.Bytes())
		}

	case 'x': // help
		io.WriteString(w, cmd.help)
	}
}

func writeFlag(b *bytes.Buffer, f string, i, n int) int {
	b.WriteString(f)
	if (i + 1) < n {
		b.WriteString(", ")
	}
	return utf8.RuneCountInString(f) + 2
}

func isShortFlag(s string) bool { return !isLongFlag(s) }
func isLongFlag(s string) bool  { return strings.HasPrefix(s, "--") }

// A CommandSet is used to construct a routing mechanism for named commands.
//
// This model is often used in CLI tools to support a list of verbs, each
// routing the caller to a sub-functionality of the main command.
//
// CommandSet satisfies the Function interface as well, making it possible to
// compose sub-commands recursively, for example:
//
//	cmd := cli.CommandSet{
//		"top": cli.CommandSet{
//			"sub-1": cli.Command(func() {
//				...
//			}),
//			"sub-2": cli.Command(func() {
//				...
//			}),
//		},
//	}
//
// The sub-commands can be called with one of these invocations:
//
//	$ program top sub-1
//
//	$ program top sub-2
//
type CommandSet map[string]Function

// Call dispatches the given arguments and environment variables to the
// sub-command named in args[0].
//
// The method returns a *Help is the first argument is -h or --help,
// and a usage error if the first argument did not match any sub-command.
//
// Call satisfies the Function interface.
func (cmds CommandSet) Call(args, env []string) (int, error) {
	for _, cmd := range cmds {
		if c, ok := cmd.(interface{ configure() }); ok {
			c.configure()
		}
	}

	// The command set supports one option, --help. A special case is made here
	// in case that option was given. It is permitted to pass --help=false, in
	// which case the option is skipped and a command name is expected after the
	// flag.
	var wantHelp bool

	if wantHelp, args = parseHelp(args); wantHelp {
		return 0, &Help{Cmd: cmds}
	}

	if len(args) == 0 {
		return 1, &Usage{Cmd: cmds, Err: fmt.Errorf("missing command")}
	}

	a := args[0]
	c := cmds[a]

	if c == nil {
		return 1, &Usage{Cmd: cmds, Err: fmt.Errorf("unknown command: %q", a)}
	}

	cmd := NamedCommand(a, c)

	code, err := cmd.Call(args[1:], env)
	switch e := err.(type) {
	case *Help:
		e.Cmd = cmd
	case *Usage:
		e.Cmd = cmd
	}
	return code, err
}

// Format writes a human-redable representation of cmds to w, using v as the
// formatting verb to determine which property of the command set should be
// written.
//
// The method supports the following formatting verbs:
//
//	%s	outputs the usage information of the command set
//	%v	outputs the full description of the command set
//
// Format satisfies the fmt.Formatter interface.
func (cmds CommandSet) Format(w fmt.State, v rune) {
	switch v {
	case 's':
		io.WriteString(w, "[command] [-h] [--help] ...")
	case 'v':
		io.WriteString(w, "Commands:\n")
		tw := newTabWriter(w)

		for _, cmd := range sortedMapKeys(reflect.ValueOf(cmds)) {
			cmdKey := cmd.String()
			fmt.Fprintf(tw, "  %s\t  %x\n", cmdKey, cmds[cmdKey])
		}

		tw.Flush()
		io.WriteString(w, `
Options:
  -h, --help  Show this help message
`)
	}
}

// NamedCommand constructs a command which carries the named passed as argument
// and delegate execution to cmd.
func NamedCommand(name string, cmd Function) Function {
	return namedCommand{name: name, cmd: cmd}
}

type namedCommand struct {
	name string
	cmd  Function
}

func (c namedCommand) Call(args, env []string) (int, error) {
	code, err := c.cmd.Call(args, env)
	switch e := err.(type) {
	case *Help:
		e.Cmd = NamedCommand(c.name, e.Cmd)
	case *Usage:
		e.Cmd = NamedCommand(c.name, e.Cmd)
	}
	return code, err
}

func (c namedCommand) Format(w fmt.State, v rune) {
	switch v {
	case 's':
		fmt.Fprintf(w, "%s ", c.name)
	}
	if f, ok := c.cmd.(fmt.Formatter); ok {
		f.Format(w, v)
	}
}

func (c namedCommand) Name() string {
	return c.name
}

func (c namedCommand) configure() {
	if x, ok := c.cmd.(interface{ configure() }); ok {
		x.configure()
	}
}

func nameOf(cmd Function) string {
	if x, ok := cmd.(interface{ Name() string }); ok {
		return x.Name()
	}
	return ""
}

func newTabWriter(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 0, ' ', 0)
}

func wantHelp(options map[string][]string) bool {
	if values, ok := options["--help"]; ok {
		if len(values) == 0 {
			return true
		}
		for _, v := range values {
			if v == "true" {
				return true
			}
		}
	}
	return false
}

func parseHelp(args []string) (wantHelp bool, next []string) {
	if len(args) == 0 {
		return
	}

	name, value, hasValue := splitNameValue(args[0])
	switch name {
	case "-h", "--help":
		next = args[1:]
	default:
		next = args
		return
	}

	if hasValue {
		wantHelp = value == "true"
	} else {
		wantHelp = true
	}

	return
}

func sortedMapKeys(m reflect.Value) []reflect.Value {
	keys := m.MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})
	return keys
}
