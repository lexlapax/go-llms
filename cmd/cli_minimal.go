//go:build minimal
// +build minimal

package main

import (
	"fmt"

	"github.com/alecthomas/kong"
)

// Minimal CLI structure without completion command
type MinimalCLI struct {
	Config   string `kong:"type='path',short='c',help='Config file location'"`
	Provider string `kong:"short='p',help='LLM provider to use'"`
	Verbose  bool   `kong:"short='v',help='Enable verbose output'"`
	Output   string `kong:"short='o',default='text',enum='text,json',help='Output format'"`
	Model    string `kong:"short='m',help='Model to use (overrides provider default)'"`

	Version    VersionCmd     `kong:"cmd,help='Show version information'"`
	Chat       ChatCmd        `kong:"cmd,help='Interactive chat with an LLM'"`
	Complete   CompleteOneCmd `kong:"cmd,help='One-shot text completion'"`
	Agent      AgentCmd       `kong:"cmd,help='Run an agent with tools'"`
	Structured StructuredCmd  `kong:"cmd,help='Get structured output from an LLM'"`
}

// MinimalCompleteCmd - simplified completion without kongplete
type MinimalCompleteCmd struct {
	Shell string `kong:"arg,required,enum='bash,zsh,fish',help='Shell to generate completion for'"`
}

func (c *MinimalCompleteCmd) Run(ctx *Context) error {
	switch c.Shell {
	case "bash":
		fmt.Println("# bash completion for go-llms - basic version")
		fmt.Println("complete -W 'chat complete agent structured version' go-llms")
	case "zsh":
		fmt.Println("# zsh completion for go-llms - basic version")
		fmt.Println("compdef _go-llms go-llms")
		fmt.Println("_go-llms() {")
		fmt.Println("  _arguments '1: :(chat complete agent structured version)'")
		fmt.Println("}")
	case "fish":
		fmt.Println("# fish completion for go-llms - basic version")
		fmt.Println("complete -c go-llms -f -a 'chat complete agent structured version'")
	}
	return nil
}