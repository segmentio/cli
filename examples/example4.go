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
