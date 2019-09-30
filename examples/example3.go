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
