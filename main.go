package main

import (
	"os"

	"github.com/vivekkundariya/grund/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
