// ABOUTME: Debug logging implementation that's only compiled with -tags debug build flag
// ABOUTME: Provides component-based debug logging controlled by GO_LLMS_DEBUG environment variable
//go:build debug
// +build debug

// Package debug provides conditional debug logging that is only compiled
// when the -tags debug build flag is used.
package debug

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	// logger is the debug logger instance
	logger = log.New(os.Stderr, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)

	// EnabledComponents tracks which components have debug logging enabled.
	// Can be set via GO_LLMS_DEBUG environment variable.
	// Format: GO_LLMS_DEBUG=all or GO_LLMS_DEBUG=param_cache,schema,agent
	EnabledComponents = make(map[string]bool)
)

func init() {
	// Parse GO_LLMS_DEBUG environment variable
	// Format: GO_LLMS_DEBUG=all or GO_LLMS_DEBUG=param_cache,schema,agent
	debugEnv := os.Getenv("GO_LLMS_DEBUG")
	if debugEnv != "" {
		if debugEnv == "all" || debugEnv == "*" {
			// Enable all components
			EnabledComponents["*"] = true
		} else {
			// Enable specific components
			components := strings.Split(debugEnv, ",")
			for _, comp := range components {
				EnabledComponents[strings.TrimSpace(comp)] = true
			}
		}
	}
}

// Printf logs a debug message if debug mode is enabled for the component
func Printf(component, format string, args ...interface{}) {
	if !isEnabled(component) {
		return
	}

	msg := fmt.Sprintf("[%s] %s", component, format)
	logger.Output(2, fmt.Sprintf(msg, args...))
}

// Println logs a debug message if debug mode is enabled for the component
func Println(component string, args ...interface{}) {
	if !isEnabled(component) {
		return
	}

	// Prepend component name to the message
	allArgs := make([]interface{}, 0, len(args)+1)
	allArgs = append(allArgs, fmt.Sprintf("[%s]", component))
	allArgs = append(allArgs, args...)

	logger.Output(2, fmt.Sprintln(allArgs...))
}

// isEnabled checks if debug logging is enabled for a component
func isEnabled(component string) bool {
	// Check if all components are enabled
	if EnabledComponents["*"] || EnabledComponents["all"] {
		return true
	}

	// Check if specific component is enabled
	return EnabledComponents[component]
}

// SetLogger allows replacing the default debug logger
func SetLogger(l *log.Logger) {
	logger = l
}
