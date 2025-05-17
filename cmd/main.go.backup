package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/willabides/kongplete"
)

var version = "v0.2.0"

func main() {
	// Parse CLI arguments
	cli := CLI{}
	parser := kong.Must(&cli,
		kong.Name("go-llms"),
		kong.Description("A Go library for LLM-powered applications with structured outputs"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	// Setup shell completion
	kongplete.Complete(parser)

	// Parse arguments
	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// The Kong library will handle the version flag automatically

	// Initialize configuration
	if err := InitConfig(cli.Config); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Load configuration into struct
	var config Config
	if err := k.Unmarshal("", &config); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config: %v\n", err)
		os.Exit(1)
	}

	// Create command context
	cmdCtx := NewContext(&cli, &config)

	// Execute command
	if err := ctx.Run(cmdCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}