package main

import (
	"fmt"
	"os"

	"github.com/nox/tnmanage/cmd"
)

func main() {
	// Load configuration from file
	_ = cmd.LoadConfig()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
