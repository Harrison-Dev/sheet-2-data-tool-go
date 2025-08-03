package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"excel-schema-generator/cmd/cli/commands"
	"excel-schema-generator/internal/adapters/excel"
	"excel-schema-generator/internal/adapters/filesystem"
	"excel-schema-generator/internal/core/data"
	"excel-schema-generator/internal/core/schema"
	"excel-schema-generator/internal/utils/errors"
	loggerAdapter "excel-schema-generator/internal/utils/logger"
	"excel-schema-generator/pkg/logger"
)

// Command represents a CLI command
type Command interface {
	Name() string
	Description() string
	SetupFlags(fs *flag.FlagSet)
	Execute(ctx context.Context, args []string) error
}

// CLI represents the command line interface
type CLI struct {
	commands map[string]Command
	logger   *logger.Logger
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	return &CLI{
		commands: make(map[string]Command),
	}
}

// AddCommand adds a command to the CLI
func (c *CLI) AddCommand(cmd Command) {
	c.commands[cmd.Name()] = cmd
}

// Run runs the CLI with the provided arguments
func (c *CLI) Run(ctx context.Context, args []string) error {
	if len(args) < 2 {
		c.printUsage()
		return errors.NewValidationError(errors.ValidationRequiredFieldCode, "No command specified")
	}

	commandName := args[1]
	cmd, exists := c.commands[commandName]
	if !exists {
		c.printUsage()
		return errors.NewValidationError(errors.ValidationInvalidValueCode, fmt.Sprintf("Unknown command: %s", commandName))
	}

	// Setup flags for the command
	fs := flag.NewFlagSet(commandName, flag.ExitOnError)
	cmd.SetupFlags(fs)

	// Parse remaining arguments
	if err := fs.Parse(args[2:]); err != nil {
		return errors.WrapError(err, errors.ValidationErrorType, errors.ValidationInvalidValueCode, "Failed to parse command arguments")
	}

	// Execute the command
	return cmd.Execute(ctx, fs.Args())
}

// printUsage prints usage information
func (c *CLI) printUsage() {
	fmt.Println("Excel Schema Generator CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  excel-schema-generator <command> [flags]")
	fmt.Println()
	fmt.Println("Available commands:")
	
	for name, cmd := range c.commands {
		fmt.Printf("  %-10s %s\n", name, cmd.Description())
	}
	
	fmt.Println()
	fmt.Println("Use 'excel-schema-generator <command> -h' for more information about a command.")
}

func main() {
	fmt.Println("=== Excel Schema Generator v0.0.9-debug ===")
	
	// Setup logging
	logConfig := logger.Config{
		Level:  slog.LevelInfo,
		Format: "text",
		Output: os.Stdout,
	}

	// Parse global flags for logging configuration
	var verbose bool
	var logLevel string
	var logFormat string

	// Create a temporary flag set just for global flags
	globalFlags := flag.NewFlagSet("global", flag.ContinueOnError)
	globalFlags.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	globalFlags.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	globalFlags.StringVar(&logFormat, "log-format", "text", "Log format (text, json)")

	// Try to parse global flags (ignore errors since they might be command-specific)
	globalFlags.Parse(os.Args[1:])

	// Update log config based on flags
	if verbose {
		logConfig.Level = slog.LevelDebug
	} else {
		logConfig.Level = parseLogLevel(logLevel)
	}
	logConfig.Format = logFormat

	// Initialize logger
	appLogger := logger.New(logConfig)
	logger.SetDefault(appLogger)

	// Create logger adapter for ports
	loggerSvc := loggerAdapter.NewLoggerAdapter(appLogger).(*loggerAdapter.LoggerAdapter)

	// Create dependencies
	fileRepo := filesystem.NewFileRepository(loggerSvc)
	excelRepo := excel.NewExcelRepository(loggerSvc)
	schemaRepo := filesystem.NewSchemaRepository(fileRepo, loggerSvc)
	outputRepo := filesystem.NewOutputRepository(fileRepo, loggerSvc)
	
	// Create error handler
	errorHandler := errors.NewErrorHandler(loggerSvc)

	// Create services
	// Note: This is a simplified setup. In a real implementation,
	// you'd want to use dependency injection container
	schemaGenerator := schema.NewSchemaGenerator(excelRepo, fileRepo, loggerSvc, nil) // validator will be nil for now
	dataGenerator := data.NewDataGenerator(excelRepo, loggerSvc, nil) // validator will be nil for now

	// Create CLI
	cli := NewCLI()
	cli.logger = appLogger

	// Add commands
	cli.AddCommand(commands.NewGenerateCommand(schemaGenerator, schemaRepo, loggerSvc))
	cli.AddCommand(commands.NewUpdateCommand(schemaGenerator, schemaRepo, loggerSvc))
	cli.AddCommand(commands.NewDataCommand(dataGenerator, schemaRepo, outputRepo, loggerSvc))

	// Create context
	ctx := context.Background()

	// Run CLI
	err := cli.Run(ctx, os.Args)
	if err != nil {
		// Handle error
		if handledErr := errorHandler.Handle(ctx, err); handledErr != nil {
			// Format user-friendly error message
			userMsg := errors.FormatUserFriendlyMessage(handledErr)
			fmt.Fprintf(os.Stderr, "Error: %s\n", userMsg)
			
			// Log detailed error for debugging
			appLogger.Error("Command execution failed", "error", handledErr)
			
			os.Exit(1)
		}
	}
}

// parseLogLevel parses log level string to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}