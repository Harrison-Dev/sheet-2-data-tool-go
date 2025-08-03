package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"excel-schema-generator/cmd/gui/app"
	"excel-schema-generator/internal/adapters/excel"
	"excel-schema-generator/internal/adapters/filesystem"
	"excel-schema-generator/internal/core/data"
	"excel-schema-generator/internal/core/models"
	"excel-schema-generator/internal/core/schema"
	"excel-schema-generator/internal/utils/errors"
	loggerAdapter "excel-schema-generator/internal/utils/logger"
	"excel-schema-generator/internal/utils/validation"
	"excel-schema-generator/pkg/logger"
)

const (
	AppName        = "Excel Schema Generator"
	AppVersion     = "0.1.0"
	schemaFileName = "schema.yml"
	dataFileName   = "output.json"
)

func main() {
	// Check if any arguments were provided
	if len(os.Args) < 2 {
		// No arguments, run GUI mode
		runGUI()
		return
	}

	// Arguments provided, run CLI mode
	runCLI()
}

// runCLI runs the application in CLI mode
func runCLI() {
	// Setup logging with default configuration
	logConfig := logger.DefaultConfig()
	appLogger := logger.New(logConfig)
	logger.SetDefault(appLogger)

	// Create logger adapter for ports
	loggerSvc := loggerAdapter.NewLoggerAdapter(appLogger).(*loggerAdapter.LoggerAdapter)

	// Create repositories
	fileRepo := filesystem.NewFileRepository(loggerSvc)
	excelRepo := excel.NewExcelRepository(loggerSvc)
	schemaRepo := filesystem.NewSchemaRepository(fileRepo, loggerSvc)
	outputRepo := filesystem.NewOutputRepository(fileRepo, loggerSvc)

	// Create services
	validator := validation.NewValidationService(loggerSvc)
	schemaGenerator := schema.NewSchemaGenerator(excelRepo, fileRepo, loggerSvc, validator)
	dataGenerator := data.NewDataGenerator(excelRepo, loggerSvc, validator)

	// Create error handler
	errorHandler := errors.NewErrorHandler(loggerSvc)

	// Create CLI application
	cli := &CLIApp{
		logger:          loggerSvc,
		errorHandler:    errorHandler,
		schemaGenerator: schemaGenerator,
		dataGenerator:   dataGenerator,
		schemaRepo:      schemaRepo,
		outputRepo:      outputRepo,
		fileRepo:        fileRepo,
	}

	// Create context
	ctx := context.Background()

	// Run CLI
	if err := cli.Run(ctx, os.Args); err != nil {
		handleError(errorHandler, loggerSvc, err)
		os.Exit(1)
	}
}

// runGUI runs the application in GUI mode
func runGUI() {
	// Setup logging with default configuration
	logConfig := logger.DefaultConfig()
	appLogger := logger.New(logConfig)
	logger.SetDefault(appLogger)

	// Create logger adapter for ports
	loggerSvc := loggerAdapter.NewLoggerAdapter(appLogger).(*loggerAdapter.LoggerAdapter)

	// Create repositories
	fileRepo := filesystem.NewFileRepository(loggerSvc)
	excelRepo := excel.NewExcelRepository(loggerSvc)

	// Create services
	validator := validation.NewValidationService(loggerSvc)
	schemaGenerator := schema.NewSchemaGenerator(excelRepo, fileRepo, loggerSvc, validator)

	// Create error handler
	errorHandler := errors.NewErrorHandler(loggerSvc)

	// Create GUI application
	guiApp := app.NewGUIApp(AppName, AppVersion, appLogger)
	guiApp.SetDependencies(schemaGenerator, fileRepo, errorHandler)

	// Run GUI
	if err := guiApp.Run(); err != nil {
		handleError(errorHandler, loggerSvc, err)
		os.Exit(1)
	}
}

// handleError handles application errors
func handleError(errorHandler *errors.ErrorHandler, logger *loggerAdapter.LoggerAdapter, err error) {
	ctx := context.Background()

	if handledErr := errorHandler.Handle(ctx, err); handledErr != nil {
		// Format user-friendly error message
		userMsg := errors.FormatUserFriendlyMessage(handledErr)
		fmt.Fprintf(os.Stderr, "Error: %s\n", userMsg)

		// Log detailed error for debugging
		logger.Error("Application error", "error", handledErr)
	}
}

// CLIApp represents the CLI application
type CLIApp struct {
	logger          *loggerAdapter.LoggerAdapter
	errorHandler    *errors.ErrorHandler
	schemaGenerator *schema.SchemaGenerator
	dataGenerator   *data.DataGenerator
	schemaRepo      *filesystem.SchemaRepository
	outputRepo      *filesystem.OutputRepository
	fileRepo        *filesystem.FileRepository
}

// Run runs the CLI application
func (app *CLIApp) Run(ctx context.Context, args []string) error {
	if len(args) < 2 {
		app.printUsage()
		return nil
	}

	// Parse common flags
	var folderPath, outputPath string
	var verbose bool

	// Simple flag parsing - in a real implementation, use flag package properly
	for i, arg := range args[2:] {
		switch {
		case arg == "-folder" && i+1 < len(args[2:]):
			folderPath = args[2:][i+1]
		case arg == "-output" && i+1 < len(args[2:]):
			outputPath = args[2:][i+1]
		case arg == "-verbose":
			verbose = true
		}
	}

	// Update logging if needed
	if verbose {
		app.logger.Debug("Verbose logging enabled")
	}

	// Validate required folder path
	if folderPath == "" {
		return errors.NewValidationError(errors.ValidationRequiredFieldCode, "Folder path is required. Use -folder flag.")
	}

	commandName := args[1]
	switch commandName {
	case "generate":
		return app.generateSchema(ctx, folderPath, outputPath)
	case "data":
		return app.generateData(ctx, folderPath, outputPath)
	default:
		app.printUsage()
		return errors.NewValidationError(errors.ValidationInvalidValueCode, fmt.Sprintf("Unknown command: %s", commandName))
	}
}

// generateSchema handles schema generation (will create new or update existing)
func (app *CLIApp) generateSchema(ctx context.Context, folderPath, outputPath string) error {
	app.logger.Info("Starting schema generation", "folder", folderPath, "output", outputPath)

	// Determine schema path
	schemaPath := app.getSchemaOutputPath(outputPath)

	// Check if schema already exists
	exists, err := app.schemaRepo.Exists(ctx, schemaPath)
	if err != nil {
		return err
	}

	var schema *models.SchemaInfo

	if exists {
		// Schema exists, perform update
		app.logger.Info("Existing schema found, updating", "path", schemaPath)
		
		// Load existing schema
		schema, err = app.schemaRepo.Load(ctx, schemaPath)
		if err != nil {
			app.logger.Error("Failed to load existing schema", "path", schemaPath, "error", err)
			return err
		}

		// Update schema
		if err := app.schemaGenerator.UpdateFromFolder(ctx, schema, folderPath); err != nil {
			app.logger.Error("Failed to update schema", "error", err)
			return err
		}
		
		fmt.Printf("Schema updated successfully: %s\n", schemaPath)
	} else {
		// Schema doesn't exist, create new
		app.logger.Info("No existing schema found, creating new", "path", schemaPath)
		
		// Generate new schema
		schema, err = app.schemaGenerator.GenerateFromFolder(ctx, folderPath)
		if err != nil {
			app.logger.Error("Failed to generate schema", "error", err)
			return err
		}
		
		fmt.Printf("Schema generated successfully: %s\n", schemaPath)
	}

	// Ensure output directory exists
	if err := app.ensureOutputDirectory(schemaPath); err != nil {
		return err
	}

	// Save schema
	if err := app.schemaRepo.Save(ctx, schema, schemaPath); err != nil {
		app.logger.Error("Failed to save schema", "path", schemaPath, "error", err)
		return err
	}

	// Success message
	fmt.Printf("Files processed: %d\n", len(schema.Files))
	fmt.Printf("Sheets found: %d\n", schema.GetSheetCount())
	app.logger.Info("Schema generation completed", "path", schemaPath, "files", len(schema.Files))

	return nil
}

// updateSchema handles schema updates
func (app *CLIApp) updateSchema(ctx context.Context, folderPath, outputPath string) error {
	app.logger.Info("Starting schema update", "folder", folderPath, "output", outputPath)

	// Determine schema path
	schemaPath := app.getSchemaOutputPath(outputPath)

	// Check if schema file exists
	exists, err := app.schemaRepo.Exists(ctx, schemaPath)
	if err != nil {
		return err
	}
	if !exists {
		return errors.NewSchemaError(errors.FileNotFoundCode, fmt.Sprintf("Schema file not found: %s. Use 'generate' command to create a new schema.", schemaPath))
	}

	// Load existing schema
	schema, err := app.schemaRepo.Load(ctx, schemaPath)
	if err != nil {
		app.logger.Error("Failed to load schema", "path", schemaPath, "error", err)
		return err
	}

	// Update schema with new data
	if err := app.schemaGenerator.UpdateFromFolder(ctx, schema, folderPath); err != nil {
		app.logger.Error("Failed to update schema", "error", err)
		return err
	}

	// Save updated schema
	if err := app.schemaRepo.Save(ctx, schema, schemaPath); err != nil {
		app.logger.Error("Failed to save updated schema", "path", schemaPath, "error", err)
		return err
	}

	// Success message
	fmt.Printf("Schema updated successfully: %s\n", schemaPath)
	fmt.Printf("Files processed: %d\n", len(schema.Files))
	fmt.Printf("Sheets found: %d\n", schema.GetSheetCount())
	app.logger.Info("Schema update completed", "path", schemaPath, "files", len(schema.Files))

	return nil
}

// generateData handles data generation (placeholder for now)
func (app *CLIApp) generateData(ctx context.Context, folderPath, outputPath string) error {
	app.logger.Info("Starting data generation", "folder", folderPath, "output", outputPath)

	// Determine schema path
	schemaPath := app.getSchemaOutputPath(outputPath)

	// Check if schema file exists
	exists, err := app.schemaRepo.Exists(ctx, schemaPath)
	if err != nil {
		return err
	}
	if !exists {
		return errors.NewSchemaError(errors.FileNotFoundCode, fmt.Sprintf("Schema file not found: %s. Use 'generate' command to create a schema first.", schemaPath))
	}

	// Load schema
	schema, err := app.schemaRepo.Load(ctx, schemaPath)
	if err != nil {
		app.logger.Error("Failed to load schema", "path", schemaPath, "error", err)
		return err
	}

	// Generate data from schema using DataGenerator
	outputData, err := app.dataGenerator.GenerateFromSchema(ctx, schema, folderPath)
	if err != nil {
		app.logger.Error("Failed to generate data", "error", err)
		return err
	}

	// Determine output path
	dataPath := app.getDataOutputPath(outputPath)

	// Ensure output directory exists
	if err := app.ensureOutputDirectory(dataPath); err != nil {
		return err
	}

	// Save output data
	if err := app.outputRepo.SaveJSON(ctx, outputData, dataPath); err != nil {
		app.logger.Error("Failed to save output data", "path", dataPath, "error", err)
		return err
	}

	// Success message
	fmt.Printf("Data generated successfully: %s\n", dataPath)
	fmt.Printf("Classes: %d\n", outputData.GetClassCount())
	fmt.Printf("Records: %d\n", outputData.GetTotalRecordCount())
	app.logger.Info("Data generation completed", "path", dataPath, "classes", outputData.GetClassCount())

	return nil
}

// getSchemaOutputPath determines the output path for the schema file
func (app *CLIApp) getSchemaOutputPath(outputPath string) string {
	if outputPath == "" {
		return schemaFileName
	}
	return filepath.Join(outputPath, schemaFileName)
}

// getDataOutputPath determines the output path for the data file
func (app *CLIApp) getDataOutputPath(outputPath string) string {
	if outputPath == "" {
		return dataFileName
	}
	return filepath.Join(outputPath, dataFileName)
}

// ensureOutputDirectory ensures the output directory exists
func (app *CLIApp) ensureOutputDirectory(outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	if outputDir == "." {
		return nil // Current directory, no need to create
	}

	ctx := context.Background()
	return app.fileRepo.CreateDir(ctx, outputDir, 0755)
}

// printUsage prints CLI usage information
func (app *CLIApp) printUsage() {
	fmt.Printf("%s v%s\n", AppName, AppVersion)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  excel-schema-generator <command> [flags]")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  generate   Generate or update schema from Excel files (auto-detects existing schema)")
	fmt.Println("  data       Generate JSON data from Excel files using an existing schema")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -folder string      Path to the Excel files folder (required)")
	fmt.Println("  -output string      Path to the output directory (optional)")
	fmt.Println("  -verbose            Enable verbose logging")
	fmt.Println("  -log-level string   Log level (debug, info, warn, error)")
	fmt.Println("  -log-format string  Log format (text, json)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  excel-schema-generator generate -folder ./excel-files")
	fmt.Println("  excel-schema-generator generate -folder ./excel-files -output ./schemas")
	fmt.Println("  excel-schema-generator data -folder ./excel-files")
	fmt.Println()
	fmt.Println("Run without arguments to start the GUI (coming soon).")
}