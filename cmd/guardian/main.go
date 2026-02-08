// Package main is the entry point for the Guardian CLI tool.
package main

import (
	"os"

	"github.com/AlexGladkov/guardian-cli/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
