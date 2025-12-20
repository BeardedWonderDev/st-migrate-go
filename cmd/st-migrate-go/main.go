package main

import (
	"fmt"
	"io"
	"os"
)

var exit = os.Exit

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	root := newRootCmd(stdout)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return err
	}
	return nil
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		exit(1)
	}
}
