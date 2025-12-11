package main

import (
	"fmt"
	"os"
)

func main() {
	root := newRootCmd(os.Stdout)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
