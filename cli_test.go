package cli_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/segmentio/cli"
)

func ExampleCommand_bool() {
	ctx := context.TODO()

	type config struct {
		Bool bool `flag:"-f,--flag"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Bool)
	})

	cli.Call(ctx, cmd)
	cli.Call(ctx, cmd, "-f")
	cli.Call(ctx, cmd, "--flag")
	cli.Call(ctx, cmd, "-f=false")
	cli.Call(ctx, cmd, "--flag=false")
	cli.Call(ctx, cmd, "-f=true")
	cli.Call(ctx, cmd, "--flag=true")

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
	ctx := context.TODO()

	type config struct {
		Int int `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Int)
	})

	cli.Call(ctx, cmd)
	cli.Call(ctx, cmd, "-f=1")
	cli.Call(ctx, cmd, "--flag=2")
	cli.Call(ctx, cmd, "-f", "3")
	cli.Call(ctx, cmd, "--flag", "4")

	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
}

func ExampleCommand_uint() {
	ctx := context.TODO()

	type config struct {
		Uint uint `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Uint)
	})

	cli.Call(ctx, cmd)
	cli.Call(ctx, cmd, "-f=1")
	cli.Call(ctx, cmd, "--flag=2")
	cli.Call(ctx, cmd, "-f", "3")
	cli.Call(ctx, cmd, "--flag", "4")

	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
}

func ExampleCommand_float() {
	ctx := context.TODO()

	type config struct {
		Float float64 `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Float)
	})

	cli.Call(ctx, cmd)
	cli.Call(ctx, cmd, "-f=1")
	cli.Call(ctx, cmd, "--flag=2")
	cli.Call(ctx, cmd, "-f", "3")
	cli.Call(ctx, cmd, "--flag", "4")

	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
}

func ExampleCommand_string() {
	ctx := context.TODO()

	type config struct {
		String string `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.String)
	})

	cli.Call(ctx, cmd)

	cli.Call(ctx, cmd, "-f=")
	cli.Call(ctx, cmd, "-f=short")
	cli.Call(ctx, cmd, "-f", "")
	cli.Call(ctx, cmd, "-f", "hello world")

	cli.Call(ctx, cmd, "--flag=")
	cli.Call(ctx, cmd, "--flag=long")
	cli.Call(ctx, cmd, "--flag", "")
	cli.Call(ctx, cmd, "--flag", "hello world")

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
	ctx := context.TODO()

	type config struct {
		Duration time.Duration `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Duration)
	})

	cli.Call(ctx, cmd)
	cli.Call(ctx, cmd, "-f=1ms")
	cli.Call(ctx, cmd, "--flag=2s")
	cli.Call(ctx, cmd, "-f", "3m")
	cli.Call(ctx, cmd, "--flag", "4h")

	// Output:
	// 0s
	// 1ms
	// 2s
	// 3m0s
	// 4h0m0s
}

func ExampleCommand_time() {
	ctx := context.TODO()

	type config struct {
		Time time.Time `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Time.Unix())
	})

	cli.Call(ctx, cmd)
	cli.Call(ctx, cmd, "-f=Mon, 02 Jan 2006 15:04:05 PST")
	cli.Call(ctx, cmd, "--flag=Mon, 02 Jan 2006 15:04:05 PST")
	cli.Call(ctx, cmd, "-f", "Mon, 02 Jan 2006 15:04:05 PST")
	cli.Call(ctx, cmd, "--flag", "Mon, 02 Jan 2006 15:04:05 PST")

	// Output:
	//-62135596800
	//1136214245
	//1136214245
	//1136214245
	//1136214245
}

func ExampleCommand_slice() {
	ctx := context.TODO()

	type config struct {
		// Slice types in the configuration struct means the flag can be
		// passed multiple times.
		Input []string `flag:"-f,--flag"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Input)
	})

	cli.Call(ctx, cmd)
	cli.Call(ctx, cmd, "-f=file1", "--flag=file2", "--flag", "file3")

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
	ctx := context.TODO()

	type config struct {
		Input unmarshaler `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(string(config.Input))
	})

	cli.Call(ctx, cmd)
	cli.Call(ctx, cmd, "--flag", "hello world")

	// Output:
	//
	// hello world
}

func ExampleCommand_default() {
	ctx := context.TODO()

	type config struct {
		Path string `flag:"-p,--path" default:"file.txt" env:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Path)
	})

	cli.Call(ctx, cmd)
	// Output: file.txt
}

func ExampleCommand_required() {
	ctx := context.TODO()

	type config struct {
		Path string `flag:"-p,--path" env:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.Path)
	})

	cli.Err = os.Stdout
	cli.Call(ctx, cmd)
	// Output:
	// Usage:
	//   [options]
	//
	// Options:
	//   -h, --help         Show this help message
	//   -p, --path string
	//
	// Error:
	//   missing required option: "--path"
}

func ExampleCommand_environment() {
	ctx := context.TODO()

	type config struct {
		String string `flag:"-f,--flag" default:"-"`
	}

	cmd := cli.Command(func(config config) {
		fmt.Println(config.String)
	})

	os.Setenv("FLAG", "hello world")
	cli.Err = os.Stdout
	cli.Call(ctx, cmd)
	// Output: hello world
}

func ExampleCommand_positional_arguments() {
	ctx := context.TODO()

	type config struct{}

	cmd := cli.Command(func(config config, x, y int) {
		fmt.Println(x, y)
	})

	cli.Call(ctx, cmd, "10", "42")
	// Output: 10 42
}

func ExampleCommand_positional_arguments_slice() {
	ctx := context.TODO()

	type config struct{}

	cmd := cli.Command(func(config config, paths []string) {
		fmt.Println(paths)
	})

	cli.Call(ctx, cmd, "file1.txt", "file2.txt", "file3.txt")
	// Output: [file1.txt file2.txt file3.txt]
}

func ExampleCommand_with_sub_command() {
	ctx := context.TODO()

	type config struct{}

	cmd := cli.Command(func(config config, sub ...string) {
		fmt.Println(sub)
	})

	cli.Call(ctx, cmd, "--", "curl", "https://segment.com")
	// Output: [curl https://segment.com]
}

func ExampleCommand_context() {
	ctx := context.TODO()

	cmd := cli.Command(func(ctx context.Context) {
		fmt.Println("hello world")
	})

	cli.Call(ctx, cmd)
	// Output: hello world
}

func ExampleCommand_context_config() {
	ctx := context.TODO()

	type config struct{}

	cmd := cli.Command(func(ctx context.Context, config config) {
		fmt.Println("hello world")
	})

	cli.Call(ctx, cmd)
	// Output: hello world
}

func ExampleCommand_context_args() {
	ctx := context.TODO()

	type config struct{}

	cmd := cli.Command(func(ctx context.Context, config config, args []string) {
		fmt.Println(args)
	})

	cli.Call(ctx, cmd, "hello", "world")
	// Output: [hello world]
}

func ExampleCommandSet() {
	ctx := context.TODO()

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

	cli.Call(ctx, cmd, "help")
	cli.Call(ctx, cmd, "do", "this")
	cli.Call(ctx, cmd, "do", "that")

	// Output:
	// help
	// this
	// that
}

func ExampleCommand_help() {
	ctx := context.TODO()

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
	cli.Call(ctx, cmd, "do", "-h")

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
	ctx := context.TODO()

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
	cli.Call(ctx, cmd, "do", "-n", "abc")

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
	ctx := context.TODO()

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
	cli.Call(ctx, cmd, "do", "--help")

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
