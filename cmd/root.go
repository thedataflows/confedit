package cmd

import (
	"fmt"
	"slices"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
	"github.com/joho/godotenv"
	log "github.com/thedataflows/go-lib-log"
)

const (
	PKG_CMD  = "cmd"
	APP_NAME = "confedit"
)

type Globals struct {
	LogLevel  string `help:"Log level (trace,debug,info,warn,error)" default:"info"`
	LogFormat string `help:"Log format (console,json)" default:"console"`
	DryRun    bool   `help:"Show what would be deleted without actually editing" default:"false"`
	Config    string `short:"c" help:"Path to CUE configuration file or directory" default:"config/"`
	Schema    string `help:"Path to alternative CUE schema file. This will override the embedded schema. Use mainly for developing schema changes."`
	StateDir  string `help:"Directory for state storage. Not yet used." default:".state/"`
}

// CLI represents the main CLI structure
type CLI struct {
	Globals  `kong:"embed"`
	Version  VersionCmd  `cmd:"" help:"Show version information"`
	Apply    ApplyCmd    `cmd:"" help:"Apply configuration to target system"`
	Status   StatusCmd   `cmd:"" help:"Check configuration status on target system"`
	List     ListCmd     `cmd:"" help:"List configured targets"`
	Generate GenerateCmd `cmd:"" help:"Generate CUE config from diff between source and target of specified type"`
}

// AfterApply is called after Kong parses the CLI but before the command runs
func (cli *CLI) AfterApply(ctx *kong.Context) error {
	// Skip initialization for version command
	if ctx.Command() == "version" || slices.Contains(ctx.Args, "--help") || slices.Contains(ctx.Args, "-h") {
		return nil
	}

	// Set log level and format
	if err := log.SetGlobalLoggerLogLevel(cli.LogLevel); err != nil {
		return fmt.Errorf("set log level: %w", err)
	}

	if err := log.SetGlobalLoggerLogFormat(cli.LogFormat); err != nil {
		return fmt.Errorf("set log format: %w", err)
	}

	return nil
}

// Run executes the CLI with the given version
func Run(version string, args []string) error {
	// Optionally load .env file if it exists
	_ = godotenv.Load(
		".env",             // Current directory
		".env.local",       // Local overrides (common in web development)
		".env.development", // Development environment
	)

	var cli CLI

	parser, err := kong.New(&cli,
		kong.Name(APP_NAME),
		kong.Description("A CLI tool for editing configuration files"),
		kong.Configuration(kongyaml.Loader),
		kong.UsageOnError(),
		kong.DefaultEnvars(""),
	)
	if err != nil {
		return fmt.Errorf("create CLI parser: %w", err)
	}

	ctx, err := parser.Parse(args)
	if slices.Contains(args, "--help") || slices.Contains(args, "-h") {
		return nil
	}
	if err != nil {
		parser.FatalIfErrorf(err)
		return err
	}

	// Check if this is the version command - handle it specially without logging/config
	if ctx.Command() == "version" {
		return ctx.Run(version)
	}

	log.Logger().Info().Str(log.KEY_PKG, PKG_CMD).Str("app", ctx.Model.Name).Str("version", version).Msg("Starting application")

	return ctx.Run(ctx, &cli)
}
