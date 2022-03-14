package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/segmentio/cli"
)

func ExampleCommand_bool() {
	type config struct {
		Bool bool `flag:"-f,--flag"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Bool)
	})

	cli.Call(cmd)
	cli.Call(cmd, "-f")
	cli.Call(cmd, "--flag")
	cli.Call(cmd, "-f=false")
	cli.Call(cmd, "--flag=false")
	cli.Call(cmd, "-f=true")
	cli.Call(cmd, "--flag=true")

	// Output:
	// false
	// true
	// true
	// false
	// false
	// true
	// true
}

func ExampleCommand_int() {
	type config struct {
		Int int `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Int)
	})

	cli.Call(cmd)
	cli.Call(cmd, "-f=1")
	cli.Call(cmd, "--flag=2")
	cli.Call(cmd, "-f", "3")
	cli.Call(cmd, "--flag", "4")

	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
}

func ExampleCommand_uint() {
	type config struct {
		Uint uint `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Uint)
	})

	cli.Call(cmd)
	cli.Call(cmd, "-f=1")
	cli.Call(cmd, "--flag=2")
	cli.Call(cmd, "-f", "3")
	cli.Call(cmd, "--flag", "4")

	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
}

func ExampleCommand_float() {
	type config struct {
		Float float64 `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Float)
	})

	cli.Call(cmd)
	cli.Call(cmd, "-f=1")
	cli.Call(cmd, "--flag=2")
	cli.Call(cmd, "-f", "3")
	cli.Call(cmd, "--flag", "4")

	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
}

func ExampleCommand_string() {
	type config struct {
		String string `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.String)
	})

	cli.Call(cmd)

	cli.Call(cmd, "-f=")
	cli.Call(cmd, "-f=short")
	cli.Call(cmd, "-f", "")
	cli.Call(cmd, "-f", "hello world")

	cli.Call(cmd, "--flag=")
	cli.Call(cmd, "--flag=long")
	cli.Call(cmd, "--flag", "")
	cli.Call(cmd, "--flag", "hello world")

	// Output:
	//
	//
	// short
	//
	// hello world
	//
	// long
	//
	// hello world
}

func ExampleCommand_duration() {
	type config struct {
		Duration time.Duration `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Duration)
	})

	cli.Call(cmd)
	cli.Call(cmd, "-f=1ms")
	cli.Call(cmd, "--flag=2s")
	cli.Call(cmd, "-f", "3m")
	cli.Call(cmd, "--flag", "4h")

	// Output:
	// 0s
	// 1ms
	// 2s
	// 3m0s
	// 4h0m0s
}

func ExampleCommand_time() {
	type config struct {
		Time time.Time `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Time.Unix())
	})

	cli.Call(cmd)
	cli.Call(cmd, "-f=Mon, 02 Jan 2006 15:04:05 UTC")
	cli.Call(cmd, "--flag=Mon, 02 Jan 2006 15:04:05 UTC")
	cli.Call(cmd, "-f", "Mon, 02 Jan 2006 15:04:05 UTC")
	cli.Call(cmd, "--flag", "Mon, 02 Jan 2006 15:04:05 UTC")

	// Output:
	//-62135596800
	//1136214245
	//1136214245
	//1136214245
	//1136214245
}

func ExampleCommand_slice() {
	type config struct {
		// Slice types in the configuration struct means the flag can be
		// passed multiple times.
		Input []string `flag:"-f,--flag"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Input)
	})

	cli.Call(cmd)
	cli.Call(cmd, "-f=file1", "--flag=file2", "--flag", "file3")

	// Output:
	// []
	// [file1 file2 file3]
}

type unmarshaler []byte

func (u *unmarshaler) UnmarshalText(b []byte) error {
	*u = b
	return nil
}

func ExampleCommand_textUnmarshaler() {
	type config struct {
		Input unmarshaler `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(string(config.Input))
	})

	cli.Call(cmd)
	cli.Call(cmd, "--flag", "hello world")

	// Output:
	//
	// hello world
}

func ExampleCommand_binaryUnmarshaler() {
	type config struct {
		URL url.URL `flag:"--url" default:"http://localhost/"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.URL.String())
	})

	cli.Call(cmd)
	cli.Call(cmd, "--url", "http://www.segment.com/")

	// Output:
	//
	// http://localhost/
	// http://www.segment.com/
}

func ExampleCommand_default() {
	type config struct {
		Path string `flag:"-p,--path" default:"file.txt" env:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Path)
	})

	cli.Call(cmd)
	// Output: file.txt
}

func ExampleCommand_required() {
	type config struct {
		Path string `flag:"-p,--path" env:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Path)
	})

	cli.Err = os.Stdout
	cli.Call(cmd)
	// Output:
	// Usage:
	//   [options]
	//
	// Options:
	//   -h, --help         Show this help message
	//   -p, --path string
	//
	// Error:
	//   missing required flag: "--path"
}

func ExampleCommand_environment() {
	type config struct {
		String string `flag:"-f,--flag" default:"-"`
	}

	// If you don't specify the name using NamedCommand, it defaults
	// to the binary name.
	cmd := cli.NamedCommand("prog", cli.Command(func(config config) {
		fmt.Println(config.String)
	}))

	os.Setenv("PROG_FLAG", "hello world")
	cli.Err = os.Stdout
	cli.Call(cmd)
	// Output: hello world
}

func ExampleCommand_positional_arguments() {
	type config struct{}

	cmd := cli.Command(func(config config, x, y int) {
		fmt.Println(x, y)
	})

	cli.Call(cmd, "10", "42")
	// Output: 10 42
}

func ExampleCommand_positional_arguments_slice() {
	type config struct{}

	cmd := cli.Command(func(config config, paths []string) {
		fmt.Println(paths)
	})

	cli.Call(cmd, "file1.txt", "file2.txt", "file3.txt")
	// Output: [file1.txt file2.txt file3.txt]
}

func ExampleCommand_with_sub_command() {
	type config struct{}

	cmd := cli.Command(func(config config, sub ...string) {
		fmt.Println(sub)
	})

	cli.Call(cmd, "--", "curl", "https://segment.com")
	// Output: [curl https://segment.com]
}

func ExampleCommand_context() {
	ctx := context.Background()

	cmd := cli.Command(func(ctx context.Context) {
		if ctx == context.TODO() {
			fmt.Println("context.TODO()")
		} else {
			fmt.Println("context.Background()")
		}
	})

	cli.Call(cmd)
	cli.CallContext(ctx, cmd)
	// Output:
	// context.TODO()
	// context.Background()
}

func ExampleCommand_context_config() {
	ctx := context.TODO()

	type config struct{}

	cmd := cli.Command(func(ctx context.Context, config config) {
		fmt.Println("hello world")
	})

	cli.CallContext(ctx, cmd)
	// Output: hello world
}

func ExampleCommand_context_args() {
	ctx := context.TODO()

	type config struct{}

	cmd := cli.Command(func(ctx context.Context, config config, args []string) {
		fmt.Println(args)
	})

	cli.CallContext(ctx, cmd, "hello", "world")
	// Output: [hello world]
}

func ExampleCommandSet() {
	help := cli.Command(func() {
		fmt.Println("help")
	})

	this := cli.Command(func() {
		fmt.Println("this")
	})

	that := cli.Command(func() {
		fmt.Println("that")
	})

	cmd := cli.CommandSet{
		"help": help,
		"do": cli.CommandSet{
			"this": this,
			"that": that,
		},
	}

	cli.Call(cmd, "help")
	cli.Call(cmd, "do", "this")
	cli.Call(cmd, "do", "that")

	// Output:
	// help
	// this
	// that
}

func ExampleCommandSet_usage_text() {
	help := cli.Command(func() {
		fmt.Println("help")
	})

	doc := cli.Command(func() {
		fmt.Println("doc")
	})

	cover := cli.Command(func() {
		fmt.Println("cover")
	})

	cmd := cli.CommandSet{
		"help": help,
		"tool": cli.CommandSet{
			"_": &cli.CommandFunc{
				Help: "run specified go tool",
			},
			"cover": cover,
			"doc":   doc,
		},
	}

	cli.Err = os.Stdout
	cli.Call(cmd, "--help")

	// Output:
	// Usage:
	//   [command] [-h] [--help] ...
	//
	// Commands:
	//   help
	//   tool  run specified go tool
	//
	// Options:
	//   -h, --help  Show this help message
}

func TestCommandSetUsage(t *testing.T) {
	doc := cli.Command(func() {
		fmt.Println("doc")
	})

	cover := cli.Command(func() {
		fmt.Println("cover")
	})

	cmd := cli.CommandSet{
		"tool": cli.CommandSet{
			"_": &cli.CommandFunc{
				Help: "run specified go tool",
			},
			"cover": cover,
			"doc":   doc,
		},
	}
	var buf bytes.Buffer
	cli.Err = &buf
	cli.Call(cmd, "tool")
	want := `
Usage:
  tool [command] [-h] [--help] ...

Commands:
  cover
  doc

Options:
  -h, --help  Show this help message

Error:
  missing command


`
	if buf.String() != want {
		t.Errorf("subcommand: got\n%q\n\n want\n%q", buf.String(), want)
	}
}

func ExampleCommandSet_option_before_command() {
	type config struct {
		String string `flag:"-f,--flag" default:"-"`
	}

	sub := cli.Command(func(config config) {
		fmt.Println(config.String)
	})

	cmd := cli.CommandSet{
		"sub": sub,
	}

	cli.Call(cmd, "-f=hello", "sub")

	// Output:
	// hello
}

func ExampleCommandSet_option_after_command() {
	type config struct {
		String string `flag:"-f,--flag" default:"-"`
	}

	sub := cli.Command(func(config config) {
		fmt.Println(config.String)
	})

	cmd := cli.CommandSet{
		"sub": sub,
	}

	cli.Call(cmd, "sub", "-f=hello")

	// Output:
	// hello
}

func ExampleCommand_help() {
	type config struct {
		Path  string `flag:"--path"     help:"Path to some file" default:"file" env:"-"`
		Debug bool   `flag:"-d,--debug" help:"Enable debug mode"`
	}

	cmd := cli.CommandSet{
		"do": cli.Command(func(config config) {
			// ...
		}),
	}

	cli.Err = os.Stdout
	cli.Call(cmd, "do", "-h")

	// Output:
	// Usage:
	//   do [options]
	//
	// Options:
	//   -d, --debug        Enable debug mode
	//   -h, --help         Show this help message
	//       --path string  Path to some file (default: file)
}

func ExampleCommand_helpContext() {
	type config struct {
		Path  string `flag:"--path"     help:"Path to some file" default:"file" env:"-"`
		Debug bool   `flag:"-d,--debug" help:"Enable debug mode"`
	}

	cmd := cli.CommandSet{
		"do": cli.Command(func(ctx context.Context, config config) {
			// ...
		}),
	}

	cli.Err = os.Stdout
	cli.CallContext(context.Background(), cmd, "do", "-h")

	// Output:
	// Usage:
	//   do [options]
	//
	// Options:
	//   -d, --debug        Enable debug mode
	//   -h, --help         Show this help message
	//       --path string  Path to some file (default: file)
}

func ExampleCommand_usage() {
	type config struct {
		Count int  `flag:"-n"         help:"Number of things"  default:"1"`
		Debug bool `flag:"-d,--debug" help:"Enable debug mode"`
	}

	cmd := cli.CommandSet{
		"do": cli.Command(func(config config) {
			// ...
		}),
	}

	cli.Err = os.Stdout
	cli.Call(cmd, "do", "-n", "abc")

	// Output:
	// Usage:
	//   do [options]
	//
	// Options:
	//   -d, --debug  Enable debug mode
	//   -h, --help   Show this help message
	//   -n int       Number of things (default: 1)
	//
	// Error:
	//   decoding "-n": strconv.ParseInt: parsing "abc": invalid syntax
}

func ExampleCommandSet_help() {
	type thisConfig struct {
		_     struct{} `help:"Call this command"`
		Path  string   `flag:"-p,--path"  help:"Path to some file" default:"file" env:"-"`
		Debug bool     `flag:"-d,--debug" help:"Enable debug mode"`
	}

	type thatConfig struct {
		_     struct{} `help:"Call that command"`
		Count int      `flag:"-n"         help:"Number of things"  default:"1"`
		Debug bool     `flag:"-d,--debug" help:"Enable debug mode"`
	}

	cmd := cli.CommandSet{
		"do": cli.CommandSet{
			"this": cli.Command(func(config thisConfig) {
				// ...
			}),
			"that": cli.Command(func(config thatConfig) {
				// ...
			}),
		},
	}

	cli.Err = os.Stdout
	cli.Call(cmd, "do", "--help")

	// Output:
	// Usage:
	//   do [command] [-h] [--help] ...
	//
	// Commands:
	//   that  Call that command
	//   this  Call this command
	//
	// Options:
	//   -h, --help  Show this help message
}

func ExampleCommandSet_help2() {
	type thisConfig struct {
		_     struct{} `help:"Call this command"`
		Path  string   `flag:"-p,--path"  help:"Path to some file" default:"file" env:"-"`
		Debug bool     `flag:"-d,--debug" help:"Enable debug mode"`
	}

	type thatConfig struct {
		_     struct{} `help:"Call that command"`
		Count int      `flag:"-n"         help:"Number of things"  default:"1"`
		Debug bool     `flag:"-d,--debug" help:"Enable debug mode"`
	}

	cmd := cli.CommandSet{
		"do": cli.CommandSet{
			"this": cli.Command(func(config thisConfig) {
				// ...
			}),
			"that": cli.Command(func(config thatConfig) {
				// ...
			}),
		},
	}

	cli.Err = os.Stdout
	cli.Call(cmd, "do", "this", "-h")

	// Output:
	// Usage:
	//   do this [options]
	//
	// Options:
	//   -d, --debug        Enable debug mode
	//   -h, --help         Show this help message
	//   -p, --path string  Path to some file (default: file)
}

func ExampleCommand_spacesInFlag() {
	type config struct {
		String string `flag:"-f, --flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.String)
	})

	cli.Call(cmd)

	cli.Call(cmd, "-f=short")
	cli.Call(cmd, "--flag", "hello world")

	// Output:
	// short
	// hello world
}

func ExampleCommand_embedded_struct() {
	type embed struct {
		AnotherString string `flag:"--another-string" default:"b"`
	}

	type config struct {
		String string `flag:"--string" default:"a"`
		embed
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.String, config.AnotherString)
	})

	cli.Call(cmd)
	cli.Call(cmd, "--string", "A")
	cli.Call(cmd, "--another-string", "B")
	cli.Call(cmd, "--string", "A", "--another-string", "B")

	// Output:
	// a b
	// A b
	// a B
	// A B
}

func TestHelpFormat(t *testing.T) {
	var c cli.Help
	got := fmt.Sprintf("%#v", c)
	if want := "cli.Help{Cmd:cli.Function(nil)}"; got != want {
		// this is not going to be the most useful when it's also got format
		// strings, but probably better than nothing...
		t.Errorf("Sprintf(%%#v, cli.Help{}): got %q, want %q", got, want)
	}
}
