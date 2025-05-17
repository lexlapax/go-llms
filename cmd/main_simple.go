//go:build simple
// +build simple

package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/willabides/kongplete"
)

// Version information (set during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// cli structure remains same
type CLI_Simple struct {
	Config   string `kong:"type='path',short='c',help='Config file location'"`
	Provider string `kong:"short='p',help='LLM provider to use'"`
	Verbose  bool   `kong:"short='v',help='Enable verbose output'"`
	Output   string `kong:"short='o',default='text',enum='text,json',help='Output format'"`
	Model    string `kong:"short='m',help='Model to use (overrides provider default)'"`

	Version    VersionCmd    `kong:"cmd,help='Show version information'"`
	Completion CompleteCmd   `kong:"cmd,help='Generate shell completion script'"`
	Chat       ChatCmd       `kong:"cmd,help='Interactive chat with an LLM'"`
	Complete   CompleteOneCmd `kong:"cmd,help='One-shot text completion'"`
	Agent      AgentCmd      `kong:"cmd,help='Run an agent with tools'"`
	Structured StructuredCmd `kong:"cmd,help='Get structured output from an LLM'"`
}

// Context for passing configuration to commands
type SimpleContext struct {
	Config Config
}

func main_simple() {
	var cli CLI_Simple
	parser := kong.Must(&cli,
		kong.Name("go-llms"),
		kong.Description("CLI for interacting with various LLM providers"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
	)

	// Setup kongplete for shell completion
	kongplete.Complete(parser,
		kongplete.WithPredictor("file", kongplete.FilesPredictor(true)),
		kongplete.WithPredictor("path", kongplete.FilesPredictor(true)),
	)

	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	// Initialize configuration
	if err := LoadConfig(cli.Config); err != nil {
		parser.FatalIfErrorf(err)
	}

	// Override with CLI flags if provided
	if cli.Provider != "" {
		config.Provider = cli.Provider
	}
	if cli.Model != "" {
		config.Model = cli.Model
	}
	if cli.Verbose {
		config.Verbose = cli.Verbose
	}
	if cli.Output != "" {
		config.Output = cli.Output
	}

	// Execute the selected command
	err = ctx.Run(&SimpleContext{Config: config})
	parser.FatalIfErrorf(err)
}