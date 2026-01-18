package main

import (
	"os"

	"github.com/yourorg/grund/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
