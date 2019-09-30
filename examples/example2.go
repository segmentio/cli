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
