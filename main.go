package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"excel-schema-generator/excelschema"
	"excel-schema-generator/pkg/logger"
)

const (
	schemaFileName = "schema.yml"
	dataFileName   = "output.json"
)

func main() {
	// Define CLI flags
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
	dataCmd := flag.NewFlagSet("data", flag.ExitOnError)

	// Common flags
	var (
		folderPath string
		verbose    bool
		logLevel   string
		logFormat  string
	)

	// Single operation flags
	generateCmd.StringVar(&folderPath, "folder", "", "Path to the Excel files folder")
	generateCmd.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	generateCmd.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	generateCmd.StringVar(&logFormat, "log-format", "text", "Log format (text, json)")

	updateCmd.StringVar(&folderPath, "folder", "", "Path to the Excel files folder")
	updateCmd.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	updateCmd.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	updateCmd.StringVar(&logFormat, "log-format", "text", "Log format (text, json)")

	dataCmd.StringVar(&folderPath, "folder", "", "Path to the Excel files folder")
	dataCmd.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	dataCmd.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	dataCmd.StringVar(&logFormat, "log-format", "text", "Log format (text, json)")

	// Check if any arguments were provided
	if len(os.Args) < 2 {
		// No arguments, run GUI mode
		runGUI()
		return
	}

	// Parse the command
	switch os.Args[1] {
	case "generate":
		generateCmd.Parse(os.Args[2:])
	case "update":
		updateCmd.Parse(os.Args[2:])
	case "data":
		dataCmd.Parse(os.Args[2:])
	default:
		fmt.Println("Expected 'generate', 'update' or 'data' subcommands")
		os.Exit(1)
	}

	// Setup logging
	setupLogging(logLevel, logFormat, verbose)
	
	// Check if folder path is provided
	if folderPath == "" {
		logger.Error("Missing required parameter", "parameter", "folder")
		fmt.Println("Please provide a folder path using the -folder flag")
		os.Exit(1)
	}
	
	// Execute the appropriate command
	if generateCmd.Parsed() {
		generateBasicSchema(folderPath)
	} else if updateCmd.Parsed() {
		updateSchema(folderPath)
	} else if dataCmd.Parsed() {
		generateData(folderPath)
	}
}

func setupLogging(level, format string, verbose bool) {
	logLevel := parseLogLevel(level)
	config := logger.Config{
		Level:  logLevel,
		Format: format,
		Output: os.Stdout,
	}
	
	if verbose {
		config.Level = slog.LevelDebug
	}
	
	logger.SetDefault(logger.New(config))
}

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

func generateBasicSchema(folderPath string) {
	logger.Info("Starting schema generation", "folder", folderPath)
	
	schema, err := excelschema.GenerateBasicSchemaFromFolder(folderPath)
	if err != nil {
		logger.Error("Failed to generate schema", "folder", folderPath, "error", err)
		fmt.Printf("Error generating schema: %v\n", err)
		return
	}
	
	err = schema.SaveToFile(schemaFileName)
	if err != nil {
		logger.Error("Failed to save schema", "file", schemaFileName, "error", err)
		fmt.Printf("Error saving schema: %v\n", err)
		return
	}
	
	logger.Info("Schema generation completed", "file", schemaFileName)
	fmt.Printf("%s has been generated successfully in the current working directory\n", schemaFileName)
}

func updateSchema(folderPath string) {
	logger.Info("Starting schema update", "folder", folderPath, "schema", schemaFileName)
	
	schema, err := excelschema.LoadSchemaFromFile(schemaFileName)
	if err != nil {
		logger.Error("Failed to load schema", "file", schemaFileName, "error", err)
		fmt.Printf("Error loading schema: %v\n", err)
		return
	}
	
	err = excelschema.UpdateSchemaFromFolder(schema, folderPath)
	if err != nil {
		logger.Error("Failed to update schema", "folder", folderPath, "error", err)
		fmt.Printf("Error updating schema: %v\n", err)
		return
	}
	
	err = schema.SaveToFile(schemaFileName)
	if err != nil {
		logger.Error("Failed to save updated schema", "file", schemaFileName, "error", err)
		fmt.Printf("Error saving updated schema: %v\n", err)
		return
	}
	
	logger.Info("Schema update completed", "file", schemaFileName)
	fmt.Printf("%s has been updated successfully in the current working directory\n", schemaFileName)
}

func generateData(folderPath string) {
	logger.Info("Starting data generation", "folder", folderPath, "schema", schemaFileName)
	
	schema, err := excelschema.LoadSchemaFromFile(schemaFileName)
	if err != nil {
		logger.Error("Failed to load schema", "file", schemaFileName, "error", err)
		fmt.Printf("Error loading schema: %v\n", err)
		return
	}
	
	output, err := excelschema.GenerateDataFromFolder(schema, folderPath)
	if err != nil {
		logger.Error("Failed to generate data", "folder", folderPath, "error", err)
		fmt.Printf("Error generating data: %v\n", err)
		return
	}
	
	err = excelschema.SaveJSONOutput(output, dataFileName)
	if err != nil {
		logger.Error("Failed to save data", "file", dataFileName, "error", err)
		fmt.Printf("Error saving data: %v\n", err)
		return
	}
	
	logger.Info("Data generation completed", "file", dataFileName)
	fmt.Printf("%s has been generated successfully in the current working directory\n", dataFileName)
}
