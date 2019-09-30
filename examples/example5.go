package main

import (
	"fmt"

	"github.com/segmentio/cli"
)

func main() {
	one := cli.Command(func() {
		fmt.Println("1")
	})

	two := cli.Command(func() {
		fmt.Println("2")
	})

	three := cli.Command(func() {
		fmt.Println("3")
	})

	cli.Exec(cli.CommandSet{
		"one":   one,
		"two":   two,
		"three": three,
	})
}
