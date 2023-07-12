package cli

import (
	"bytes"
	"context"
	"errors"
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
// to the command is expected to be invoked with no arguments, for example:
//
//	cmd := cli.Command(func() {
//		...
//	})
//
// If the function accepts arguments, the first argument (except for an optional
// initial `context.Context`) must always be a struct type, which describes the
// set of options that are accepted by the command:
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
// Five keys are recognized in the struct tags: "flag", "env", "help",
// "default", and "hidden".
//
// The "flag" struct tag is a comma-separated list of command line flags that
// map to the field. This tag is required.
//
// The "env" struct tag optionally specifies the name of an environment variable
// whose value may provide a field value. When the tag is not specified, then
// environment variables corresponding to long command line flags may provide
// field values. A tag value of "-" disables this default behavior.
//
// The "help" struct tag is a human-readable message describing what the field is
// used for.
//
// The "default" struct tag provides the default value of the field when the
// argument was missing from the call to the command. Any flag which has no
// default value and isn't a boolean or a slice type must be passed when calling
// the command, otherwise a usage error is returned. The special default value
// "-" can be used to indicate that the option is not required and should assume
// its zero-value when omitted.
//
// The "hidden" struct flag is a Boolean indicating if the field should be
// excluded from help text, essentially making it undocumented.
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
// a nested function when invoked.
type CommandFunc struct {
	// A short help message describing what the command does.
	Help string

	// A full description of the command.
	Desc string

	// The function that the command calls out to when invoked.
	//
	// See Command for details about the accepted signatures.
	Func interface{}

	// An optional usage string for this function. If set, then this replaces
	// the default one that shows the types (but not names) of arguments.
	Usage string

	// Set of options to not set from the environment
	// this is a more user-friendly-syntax than IgnoreEnvOptionMap
	// However, this is strictly for user input and should not be used in the cli code
	// Please use IgnoreEnvOptionMap internally
	IgnoreEnvOptions []string

	// Set of options to not set from the environment
	// This is to convert IgnoreEnvOptions field to a map for efficient lookups
	IgnoreEnvOptionsMap map[string]struct{}

	function reflect.Value
	parser   parser
	options  structDecoder
	values   []decodeFunc
	variadic bool
	context  bool
	help     string
}

func (cmd *CommandFunc) configure() {
	if cmd.function.IsValid() {
		return // already configured
	}

	if cmd.Func == nil {
		panic(fmt.Sprintf("cli.Command: expected a function as argument but got nil (help text: %q, desc: %q)", cmd.Help, cmd.Desc))
	}
	t := reflect.TypeOf(cmd.Func)
	v := reflect.ValueOf(cmd.Func)

	if t.Kind() != reflect.Func {
		panic("cli.Command: expected a function as argument but got " + t.String())
	}

	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()

	cmd.function = v
	cmd.variadic = t.IsVariadic()

	if n := t.NumIn(); n == 0 {
		cmd.parser, cmd.options, cmd.help = makeStructDecoder(emptyType)
	} else {
		x := 0

		if f := t.In(x); f.Kind() == reflect.Interface && f.Implements(ctxType) {
			cmd.context = true
			x++
		}

		if x < n {
			if f := t.In(x); f.Kind() == reflect.Struct {
				cmd.parser, cmd.options, cmd.help = makeStructDecoder(f)
				x++
			} else {
				panic("cli.Command: expected a struct as first argument but got " + f.String())
			}
		}

		if cmd.variadic {
			n--
		}

		for i := x; i < n; i++ {
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
			panic(
				"cli.Command: expected a function returning (error) but got (" + r0.String() + ")",
			)
		}
	case 2:
		if r0, r1 := t.Out(0), t.Out(1); r0 != intType || r1 != errorType {
			panic(
				"cli.Command: expected a function returing (int, error) but got (" + r0.String() + ", " + r1.String() + ")",
			)
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
func (cmd *CommandFunc) Call(ctx context.Context, args, env []string) (int, error) {
	cmd.configure()

	options, values, command, err := cmd.parser.parseCommandLine(args)
	if err != nil {
		return 1, err
	}

	if wantHelp(options) {
		return 0, &Help{Cmd: cmd}
	}

	// If user chooses to pass in IgnoreEnvOptionsMap instead of IgnoreEnvOptions
	// we do not reset it
	if cmd.IgnoreEnvOptionsMap == nil {
		cmd.IgnoreEnvOptionsMap = make(map[string]struct{})
	}
	// Convert list to string for a faster look up
	for _, name := range cmd.IgnoreEnvOptions {
		cmd.IgnoreEnvOptionsMap[name] = struct{}{}
	}

	for name, field := range cmd.options {

		if _, ok := cmd.IgnoreEnvOptionsMap[name]; ok {
			continue
		}

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
			return 1, &Usage{Cmd: cmd, Err: fmt.Errorf("missing required flag: %q", name)}
		}
	}

	var params []reflect.Value

	x := 0

	if cmd.context {
		params = append(params, reflect.ValueOf(ctx))
		x++
	} else if ctx != nil && ctx != context.TODO() {
		panic("to use context, all commands must accept a context.Context as their first argument")
	}

	if t := cmd.function.Type(); t.NumIn() > 0 {
		// Positional arguments are decoded into each following function
		// parameter, until a slice type is encountered which receives all
		// the remaining values.
		n := t.NumIn()

		if x < n {
			// Configuration options are decoded into the first function parameter.
			v := reflect.New(t.In(x)).Elem()
			if err := cmd.options.decode(v, options); err != nil {
				if uerr, ok := err.(*Usage); ok {
					uerr.Cmd = cmd
				}
				if herr, ok := err.(*Help); ok {
					herr.Cmd = cmd
				}
				return 1, err
			}
			params = append(params, v)
			x++
		}

		if cmd.variadic {
			n--
		}

		for i := x; i < n; i++ {
			p := t.In(i)
			v := reflect.New(p).Elem()

			if p.Kind() == reflect.Slice {
				if err := cmd.values[i-x](v, values); err != nil {
					return 1, err
				}
				params = append(params, v)
				values = nil
				break
			}

			var value []string
			if len(values) == 0 {
				value = []string{""}
			} else {
				value, values = values[:1], values[1:]
			}

			if err := cmd.values[i-x](v, value); err != nil {
				return 1, err
			}
			params = append(params, v)
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

// Format satisfies the fmt.Formatter interface. It recognizes the following
// verbs:
//
//	%s	outputs the usage information of the command
//	%v	outputs the full description of the command
//	%x	outputs the help message of the command
func (cmd *CommandFunc) Format(w fmt.State, v rune) {
	switch v {
	case 's': // usage
		if cmd.Usage != "" {
			io.WriteString(w, cmd.Usage)
			return
		}

		io.WriteString(w, "[options]")

		t := cmd.function.Type()
		n := t.NumIn()
		if cmd.variadic {
			n--
		}

		i := 1
		if cmd.context {
			i = 2
		}

		for i < n {
			p := t.In(i)
			fmt.Fprintf(w, " [%s]", typeNameOf(p))

			if p.Kind() == reflect.Slice {
				break
			}

			i++
		}

		if cmd.variadic {
			io.WriteString(w, " -- [command]")
		}

	case 'v': // description
		if w.Flag('#') {
			io.WriteString(
				w,
				strings.Replace(cmd.function.Type().String(), "func", "cli.Command", 1),
			)
			return
		}

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
			if field.hidden {
				continue
			}
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
			field := cmd.options[fieldName.String()]
			if field.hidden {
				continue
			}

			b.Reset()
			b.WriteString("  ") // indent

			// This counter is used to track how many short and long flags have
			// been written.
			//
			// Short flags are printed first, then long flags. Empty columns are
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
		if cmd.help != "" {
			io.WriteString(w, cmd.help)
		} else if cmd.Help != "" {
			// if we're asking for help text, we may not have called configure()
			// on this CommandFunc yet
			io.WriteString(w, cmd.Help)
		}
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
type CommandSet map[string]Function

// Call dispatches the given arguments and environment variables to the
// sub-command named in the first non-option value in args. Finding the command
// separator "--" before a sub-command name results in an error.
//
// The method returns a *Help (as an error) is the first argument is -h or
// --help, and a usage error if the first argument did not match any
// sub-command.
//
// Call satisfies the Function interface.
func (cmds CommandSet) Call(ctx context.Context, args, env []string) (int, error) {
	for cmdKey, cmd := range cmds {
		c, canConfigure := cmd.(interface{ configure() })
		// "_" is the special key for printing help - skip it
		if canConfigure && cmdKey != "_" {
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

	var a string // command name
	var c Function

	for i, arg := range args {
		if isCommandSeparator(arg) {
			break
		}
		if isOption(arg) {
			continue
		}
		a = arg
		tmp := make([]string, 0, len(args)-1)
		tmp = append(tmp, args[:i]...)
		tmp = append(tmp, args[i+1:]...)
		args = tmp
		break
	}

	if a == "" {
		return 1, &Usage{Cmd: cmds, Err: fmt.Errorf("missing command")}
	}

	if c = cmds[a]; c == nil {
		minLevenshtein := 1000
		closestCommand := ""
		for cmd := range cmds {
			score := levenshtein(a, cmd)
			if score < minLevenshtein {
				closestCommand = cmd
				minLevenshtein = score
			}
		}
		errMessage := fmt.Sprintf("unknown command: %q", a)
		if similarEnough(a, closestCommand, minLevenshtein) {
			errMessage += fmt.Sprintf(". Did you mean %q? Use --help to see all commands",
				closestCommand)
			return 1, errors.New(errMessage)
		}
		return 1, &Usage{Cmd: cmds, Err: errors.New(errMessage)}
	}

	return NamedCommand(a, c).Call(ctx, args, env)
}

// similarEnough determines if input and want are similar enough. If input and
// want are 2 characters, we maybe don't want to issue a suggestion because
// you're changing 50% of the word. But longer words a Levenshtein distance of
// 2 is probably good.
func similarEnough(input, want string, levenshtein int) bool {
	if len(input) <= 1 || len(want) <= 1 {
		return false
	}
	if len(input) <= 3 || len(want) <= 3 {
		return levenshtein <= 1
	}
	frac := float64(levenshtein) / float64(len(want))
	// this allows 2 out of 7 letters off but forbids 2 out of 6
	return frac <= 0.3
}

// Format writes a human-readable representation of cmds to w, using v as the
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
		if w.Flag('#') {
			io.WriteString(w, "cli.CommandSet{")
			names := make([]string, 0, len(cmds))
			for name := range cmds {
				names = append(names, name)
			}
			sort.Strings(names)
			for i, name := range names {
				if i != 0 {
					io.WriteString(w, ", ")
				}
				fmt.Fprintf(w, "%q:%#v", name, cmds[name])
			}
			io.WriteString(w, "}")
			return
		}

		io.WriteString(w, "Commands:\n")
		tw := newTabWriter(w)

		for _, cmd := range sortedMapKeys(reflect.ValueOf(cmds)) {
			cmdKey := cmd.String()
			if cmdKey == "_" {
				// Short flag for help text, not a runnable command.
				continue
			}
			fmt.Fprintf(tw, "  %s", cmdKey)
			// Avoid printing the whitespace if there's no value - makes it
			// easier to write tests against with text editors that
			// strip extraneous whitespace from the ends of lines.
			val := fmt.Sprintf("%x", cmds[cmdKey])
			if val != "" {
				io.WriteString(tw, "\t  "+val)
			}
			tw.Write([]byte{'\n'})
		}

		tw.Flush()
		io.WriteString(w, `
Options:
  -h, --help  Show this help message
`)
	case 'x':
		if cmd, ok := cmds["_"]; ok {
			fmt.Fprintf(w, "%x", cmd)
		}
	}
}

// NamedCommand constructs a command which carries the name passed as argument
// and delegate execution to cmd.
func NamedCommand(name string, cmd Function) Function {
	return &namedCommand{name: name, cmd: cmd}
}

// namedCommand is a command function associated with a command name. The
// only purpose of the wrapper is to preserve the command name for output.
type namedCommand struct {
	name string
	cmd  Function
}

// Call dispatches the given arguments and environment variables to the function
// in this named sub-command.
//
// Call satisfies the Function interface.
func (c *namedCommand) Call(ctx context.Context, args, env []string) (int, error) {
	code, err := c.cmd.Call(ctx, args, env)
	switch e := err.(type) {
	case *Help:
		if e.Cmd == nil {
			e.Cmd = c
		} else {
			e.Cmd = NamedCommand(c.name, e.Cmd)
		}
	case *Usage:
		if e.Cmd == nil {
			e.Cmd = c
		} else {
			e.Cmd = NamedCommand(c.name, e.Cmd)
		}
	}
	return code, err
}

func (c *namedCommand) Format(w fmt.State, v rune) {
	switch v {
	case 's':
		fmt.Fprintf(w, "%s ", c.name)
	case 'v':
		if w.Flag('#') {
			fmt.Fprintf(w, "cli.NamedCommand(%q, %#v)", c.name, c.cmd)
			return
		}
	}
	if f, ok := c.cmd.(fmt.Formatter); ok {
		f.Format(w, v)
	}
}

// Name retrieves this command's name.
func (c *namedCommand) Name() string {
	return c.name
}

func (c *namedCommand) configure() {
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
