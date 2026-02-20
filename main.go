package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/dmitriyb/conductor/internal/config"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run parses flags and dispatches to the appropriate subcommand.
// It returns the exit code. Extracted from main() for testability.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("conductor", flag.ContinueOnError)
	fs.SetOutput(stderr)
	cfgPath := fs.String("config", "orchestrator.yaml", "config file path")
	logLevel := fs.String("log-level", "info", "log level")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	subcmds := fs.Args()
	if len(subcmds) == 0 {
		fmt.Fprintln(stderr, "usage: conductor [flags] <subcommand>")
		fmt.Fprintln(stderr, "subcommands: validate, build, run")
		return 1
	}

	switch subcmds[0] {
	case "validate", "build", "run":
		// valid subcommand — continue below
	default:
		fmt.Fprintf(stderr, "unknown subcommand: %q\n", subcmds[0])
		fmt.Fprintln(stderr, "subcommands: validate, build, run")
		return 1
	}

	logger := config.InitLogging(*logLevel, stderr)

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		return 1
	}

	if err := config.Validate(cfg); err != nil {
		logger.Error("config validation failed", "error", err)
		return 1
	}

	switch subcmds[0] {
	case "validate":
		fmt.Fprintln(stdout, "configuration is valid")
	case "build":
		fmt.Fprintln(stderr, "build: not yet implemented")
		return 1
	case "run":
		fmt.Fprintln(stderr, "run: not yet implemented")
		return 1
	}

	return 0
}
