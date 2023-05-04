# cli ![build status](https://github.com/segmentio/cli/actions/workflows/go.yml/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/segmentio/cli)](https://goreportcard.com/report/github.com/segmentio/cli) [![GoDoc](https://godoc.org/github.com/segmentio/cli?status.svg)](https://godoc.org/github.com/segmentio/cli)

Go package providing high-level constructs for command-line tools.

## Motivation

The user interface of a program is a major contributor to its adoption and
maintainability, however it is often overlooked as a second-class requirement.
Developers often focus on the _core functionalities_ of their programs and don't
put as much time in designing and understanding how the program will be used.

The reality is that even when effort is spent on building powerful interfaces,
the tooling available in Go can be a blocker to generalization of the practice.

The standard library does offer a [package](https://golang.org/pkg/flag/) for
parsing command line arguments, but it is limited to flags, and doesn't support
loading configuration options from the environment, or building advanced UX with
sub-commands.

Another popular package is [spf13/cobra](https://godoc.org/github.com/spf13/cobra),
which has been the to-go solution for most projects. This package is powerful
but also very large, brings a lot of complexity to programs that use it, and
can be very time consuming to navigate for developers.

We believed that creating powerful tools should be simple, that developers
should be empowered to build programs that are safe to use and easy to evolve.

The `segmentio/cli` package was designed to have a minimal yet flexible API,
making it easy to learn, and offering clear guidlines on how to build and evolve
command line programs.

## Command Line Interface

This section contains a couple of examples that showcase the features of the
package. (For more, see the "examples" directory.)

### Flags

This first example presents how to construct a command which accepts a --name
flag:

```go
package main

import (
	"fmt"

	"github.com/segmentio/cli"
)

func main() {
	type config struct {
		Name string `flag:"-n,--name" help:"Someone's name" default:"Luke"`
	}

	cli.Exec(cli.Command(func(config config) {
		fmt.Printf("hello %s!\n", config.Name)
	}))
}
```

```
$ ./example1 --help

Usage:
  example1 [options]

Options:
  -h, --help         Show this help message
  -n, --name string  Someone's name (default: Luke)

```

```
$ ./example1 --name Han
hello Han!
```

The key take away here is how flags are declared by the first argument of the
function implementing the command. The `segmentio/cli` package implements a
calling convention which maps the program arguments to the arguments of the
function being called.

### Default Values

The first example shows how to set a default value for a flag. If a flag is
truly optional, then set its default value to "-"; when the flag isn't used, its
field assumes its zero-value. When a flag does not have any default value
defined, then it is required.

```go
type config struct {
	// optional, default "Luke"
	Name     string `flag:"-n,--name"     help:"Someone's name"        default:"Luke"`
	// optional, no default
	Planet   string `flag:"-p,--planet"   help:"Someone's home planet" default:"-"`
	// required
	Greeting string `flag:"-g,--greeting" help:"Greeting word, such as hello"`
}
```

### Hidden Flags

A hidden flag is not included in help text, making it undocumented but still
usable.

```go
	// optional, default "Leia", hidden
	Sibling  string `flag:"-s,--sibling"  help:"Secret family member"  default:"Leia" hidden:"true"`
```

### Command Help Text

When the struct used for flags contains a field named `_`, its "help" tag
defines the command's own help message. The field type is ignored.

```go
type config struct {
	_    struct{} `help:"Greets someone from a galaxy far, far away"`
	Name string   `flag:"-n,--name" help:"Someone's name" default:"Luke"`
}
```

### Positional Arguments

While the first argument of a command must always be a struct defining the set
of accepted flags, the function may also define extra arguments which will be
loaded from positional arguments:

```go
package main

import (
	"fmt"

	"github.com/segmentio/cli"
)

func main() {
	type noflags struct{}

	cli.Exec(cli.Command(func(_ noflags, x, y int) {
		fmt.Println(x + y)
	}))
}
```
```
$ ./example2 --help

Usage:
  example2 [options] [int] [int]

Options:
  -h, --help  Show this help message
```
```
$ ./example2 1 2
3
```

The last function parameter may also be a slice which captures all remaining
positional arguments:

```go
package main

import (
	"fmt"

	"github.com/segmentio/cli"
)

func main() {
	type noflags struct{}

	cli.Exec(cli.Command(func(_ noflags, words []string) {
		for _, word := range words {
			fmt.Println(word)
		}
	}))
}
```
```
$ ./example3 --help

Usage:
  example3 [options] [string...]

Options:
  -h, --help  Show this help message

```
```
$ ./example3 hello world
hello
world
```

### Child Commands

It is common for wrapper programs to accept an arbitrary command that they
execute after performing some initializations. To reduce the risk of mixing
the program's arguments and the arguments of its child-command, a "--" separator
is employed as a delimiter between the two on the command line.

With the `segmentio/cli` package, this model is supported by adding a variadic
list of string parameters to the command:

```go
package main

import (
	"fmt"
	"strings"

	"github.com/segmentio/cli"
)

func main() {
	type noflags struct{}

	cli.Exec(cli.Command(func(_ noflags, args ...string) {
		fmt.Println("run:", strings.Join(args, " "))
	}))
}
```
```
$ ./example4 --help

Usage:
  example4 [options] -- [command]

Options:
  -h, --help  Show this help message

```
```
$ ./example4 -- echo hello world
run: echo hello world
```

### Command Sets

Advanced tools often have a set of commands in a single program, each exposing
a different feature of the tool (e.g. `git checkout`, `git commit`).

The `segmentio/cli` package supports constructing programs like these using the
`cli.CommandSet` type. The next example showcases how to construct a program
accepting three sub-commands:

```go
package main

import (
	"fmt"

	"github.com/segmentio/cli"
)

func main() {
	type oneConfig struct {
		_ struct{} `help:"Usage text for command one"`
	}
	one := cli.Command(func(cfg oneConfig) {
		fmt.Println("1")
	})

	two := cli.Command(func() {
		fmt.Println("2")
	})

	three := cli.CommandSet{
		"_": cli.CommandFunc{
			Help: "Usage text for the command three",
		},
		"four": cli.Command(func() {
			fmt.Println("4")
		}),
		"five": cli.Command(func() {
			fmt.Println("4")
		}),
	}

	cli.Exec(cli.CommandSet{
		"one":   one,
		"two":   two,
		"three": three,
	})
}
```
```
$ ./example5 --help
Usage:
  example5 [command] [-h] [--help] ...

Commands:
  one    Usage text for command one
  three  Usage text for the 'three' command
  two

Options:
  -h, --help  Show this help message

```
```
$ ./example5 one
1
```

When the command set contains a value for the key `"_"`, its function value's
"Help" value defines the command set's own help message.

## Environment Variables

While passing configuration options on the command line using flags and
positional arguments provides great UX, it is also very common to use
environment variables in configuration files like kubernetes templates.

Every _long flag_ accepted by a command (flags starting with "--") can also be
loaded from environment variables. The package maps environment variables to
flags by prefixing it with the program name and converting the flag to
upper-snake-case, for example:

    > --verbose => ${PROGRAM}_VERBOSE

## Testing Commands

Testing command line programs is often overlooked, because packages which
facilitate loading program configurations often aren't designed with ease of
testing in mind.

On the other hand, commands declared with the `segmentio/cli` package are easily
testable using the `cli.Call` function, which combined with Go's support for
testable examples, offer a great model for testing commands.

Using the first example, here is how we could write tests to validate the
behavior of the command:

```go
type config struct {
	Name string `flag:"-n,--name" help:"Someone's name" default:"Luke"`
}

var command = cli.Command(func(config config) {
	fmt.Printf("hello %s!\n", config.Name)
})
```
```go
func Example_noArguments() {
	cli.Call(command)
	// Output: hello Luke!
}

func Example_withArgument() {
	cli.Call(command, "--name", "Han")
	// Output: hello Han!
}
```

## Formatting Output

A lot of command line programs also output information to their caller, and
often need to support multiple formats to be used in different conditions
(called by an operator, used in a script for automation, etc...).

This formatting work is often tedious and redundant, so the `segmentio/cli`
package exposes abstractions to help developers build tools which support
multiple output formats:

```go
type config struct {
	Output string `flag:"-o,--output" help:"Output format of the command" default:"text"`
}

type result struct {
	Name  string
	Value int
}

var command = cli.Command(func(config config) error {
	p, err := cli.Format(config.Output, os.Stdout)
	if err != nil {
		return err
	}
	defer p.Flush()

	...

	// Call p.Print one or more times to output content to stdout
	//
	// p.Print(v)
})
```

The package supports three formats out-of-the-box: text, json, and yaml.

In the text format, struct and map values are printed as table representations
with a header being the name of the struct fields or the keys of the map.
Other value types are simply printed one value per line.

All formats interpret the `json` struct tag to configure the names of the fields
and the behavior of the formatting operation.

The text format also interprets `fmt` tags as carrying the formatting string
passed in calls to functions of the `fmt` package.
