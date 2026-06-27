package main

import (
	"fmt"
	"os"

	"github.com/xhd2015/git-hooks/run"
)

func main() {
	if err := run.Main(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
