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
