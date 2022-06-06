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
		"_": &cli.CommandFunc{
			Help: "Usage text for the 'three' command",
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
