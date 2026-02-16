// Package main implements the goagent CLI tool for building and running AI agents.
package main

import (
	"context"
	"flag"
	"fmt"
)

var version = "dev" // set via ldflags during build

func main() {
	var (
		showHelp    = flag.Bool("help", false, "show help message")
		showVersion = flag.Bool("version", false, "show version")
	)
	flag.Parse()

	if *showHelp {
		showUsage()
		return
	}

	if *showVersion {
		fmt.Printf("goagent %s\n", version)
		return
	}

	// Default behavior - show help if no flags provided
	if flag.NArg() == 0 {
		showUsage()
		return
	}

	// For now, just acknowledge the command
	fmt.Printf("goagent %s - AI Agent SDK for Go\n", version)
	fmt.Println("Agent execution not yet implemented")

	ctx := context.Background()
	_ = ctx // Will be used in future implementation
}

func showUsage() {
	fmt.Printf(`goagent %s - AI Agent SDK for Go

USAGE:
    goagent [options] [command]

OPTIONS:
    --help       Show this help message
    --version    Show version information

COMMANDS:
    run          Execute an agent (not yet implemented)

EXAMPLES:
    goagent --help
    goagent --version
    goagent run

For more information, visit: https://github.com/oskarhane/goagent
`, version)
}
