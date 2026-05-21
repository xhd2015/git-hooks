package main

import (
	"fmt"
	"os"

	"git-hooks/run"
)

func main() {
	if err := run.Main(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
