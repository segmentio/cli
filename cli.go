// Package cli provides high-level tools for building command-line interfaces.
package cli

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Err is used by the Exec and Call functions to print out errors returned by
// the commands they call out to.
var Err io.Writer = os.Stderr

// The Function interface is implemented by commands that may be invoked with
// argument and environment variable lists.
//
// Functions returns a status code intended to be the exit code of the program
// that called them as well as a non-nil error if the function call failed.
type Function interface {
	Call(ctx context.Context, args, env []string) (int, error)
}

// Exec delegate the program execution to cmd, then exits with the code returned
// by the function call.
//
// A typical use case is for Exec to be the last statement of the main function
// of a program:
//
//	func main() {
//		cli.Exec(cli.Command(func(config config) {
//			...
//		})
//	}
//
// The Exec function never returns.
func Exec(cmd Function) {
	ExecContext(context.TODO(), cmd)
}

// ExecContext calls Exec but with a specified context.Context.
func ExecContext(ctx context.Context, cmd Function) {
	name := filepath.Base(os.Args[0])
	args := os.Args[1:]
	prog := NamedCommand(name, cmd)
	os.Exit(CallContext(ctx, prog, args...))
}

// Call calls cmd with args and environment variables prefixed with the
// uppercased program name.
//
// This function is often used to test commands in example programs with
// constructs like:
//
//	var command = cli.Command(func(config config) {
//		...
//	})
//
// Then in the test file:
//
//	func Example_command_with_option() {
//		cli.Call(command, "--option", "value")
//		// Output:
//		// ...
//	}
func Call(cmd Function, args ...string) int {
	return CallContext(context.TODO(), cmd, args...)
}

// CallContext calls Call but with a specified context.Context.
func CallContext(ctx context.Context, cmd Function, args ...string) int {
	prefix := strings.ToUpper(snakecase(nameOf(cmd)))
	if prefix != "" {
		prefix = prefix + "_"
	}

	code, err := cmd.Call(ctx, args, environ(prefix))

	switch err.(type) {
	case nil:
	case *Help, *Usage:
		fmt.Fprintln(Err, err)
	default:
		if err != nil {
			errorLogger := log.New(Err, "", log.LstdFlags)
			errorLogger.Print(err)
			code = 1
		}
	}

	return code
}

func environ(prefix string) []string {
	env := os.Environ()
	ret := make([]string, 0, len(env))

	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			ret = append(ret, strings.TrimPrefix(e, prefix))
		}
	}

	return ret
}

// Help values are returned by commands to indicate to the caller that it was
// called with a configuration that requested a help message rather than
// executing the command. This type satisfies the error interface.
type Help struct {
	Cmd Function
}

// Fallback for unimplemented fmt verbs
type fmtHelp struct{ cmd Function }

// Error satisfies the error interface.
func (h *Help) Error() string { return fmt.Sprintf("help: %s", h.Cmd) }

// Format satisfies the fmt.Formatter interface, print the help message for the
// command carried by h.
func (h *Help) Format(w fmt.State, v rune) {
	switch v {
	case 's':
		printUsage(w, h.Cmd)
		printHelp(w, h.Cmd)
	case 'v':
		if w.Flag('#') {
			io.WriteString(w, "cli.Help{")
			fmt.Fprintf(w, "%#v", h.Cmd)
			io.WriteString(w, "}")
			return
		}
		printUsage(w, h.Cmd)
		printHelp(w, h.Cmd)
	default:
		// fall back to default struct formatter. TODO this does not handle
		// flags
		fmt.Fprintf(w, "%"+string(v), fmtHelp{h.Cmd})
	}
}

// Usage values are returned by commands to indicate that the combination of
// arguments and environment variables they were called with was invalid. This
// type satisfies the error interface.
type Usage struct {
	Cmd Function
	Err error
}

// Error satisfies the error interface.
func (u *Usage) Error() string { return fmt.Sprintf("usage: %s: %s", u.Cmd, u.Err) }

// Format satisfies the fmt.Formatter interface, print the usage message for the
// command carried by u.
func (u *Usage) Format(w fmt.State, v rune) {
	if v == 'v' && w.Flag('#') {
		io.WriteString(w, "cli.Usage{")
		fmt.Fprintf(w, "Cmd: %#v, ", u.Cmd)
		fmt.Fprintf(w, "Err: %#v", u.Err)
		io.WriteString(w, "}")
		return
	}
	// TODO: better detection/printing based on the requested format string.
	if u.Cmd != nil {
		printUsage(w, u.Cmd)
		printHelp(w, u.Cmd)
	}
	if u.Err != nil {
		printError(w, u.Err)
	}
}

// Unwrap satisfies the errors wrapper interface.
func (u *Usage) Unwrap() error { return u.Err }

func printUsage(w io.Writer, cmd Function) {
	const format = `
Usage:
  %s

`
	fmt.Fprintf(w, format, cmd)
}

func printHelp(w io.Writer, cmd Function) {
	fmt.Fprintf(w, "%v", cmd)
}

func printError(w io.Writer, err error) {
	const format = `
Error:
  %s

`
	fmt.Fprintf(w, format, err)
}
