package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"excel-schema-generator/cmd/cli/flags"
	"excel-schema-generator/internal/ports"
	"excel-schema-generator/internal/utils/errors"
)

// DataCommand implements the generate data command
type DataCommand struct {
	dataService   ports.DataService
	schemaRepo    ports.SchemaRepository
	outputRepo    ports.OutputRepository
	logger        ports.LoggingService
	flags         flags.CommonFlags
}

// NewDataCommand creates a new data command
func NewDataCommand(
	dataService ports.DataService,
	schemaRepo ports.SchemaRepository,
	outputRepo ports.OutputRepository,
	logger ports.LoggingService,
) *DataCommand {
	return &DataCommand{
		dataService: dataService,
		schemaRepo:  schemaRepo,
		outputRepo:  outputRepo,
		logger:      logger,
	}
}

// Name returns the command name
func (c *DataCommand) Name() string {
	return "data"
}

// Description returns the command description
func (c *DataCommand) Description() string {
	return "Generate JSON data from Excel files using an existing schema"
}

// SetupFlags sets up command-specific flags
func (c *DataCommand) SetupFlags(fs *flag.FlagSet) {
	flags.AddCommonFlags(fs, &c.flags)
}

// Execute executes the data generation command
func (c *DataCommand) Execute(ctx context.Context, args []string) error {
	c.logger.Info("Starting data generation command", 
		"folder", c.flags.FolderPath, 
		"output", c.flags.OutputPath)

	// Validate flags
	if err := c.flags.Validate(); err != nil {
		return errors.WrapError(err, errors.ValidationErrorType, errors.ValidationRequiredFieldCode, "Invalid command flags")
	}

	// Determine schema path
	schemaPath := c.getSchemaPath()

	// Check if schema file exists
	exists, err := c.schemaRepo.Exists(ctx, schemaPath)
	if err != nil {
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to check schema file existence")
	}
	if !exists {
		return errors.NewSchemaError(errors.FileNotFoundCode, fmt.Sprintf("Schema file not found: %s. Use 'generate' command to create a schema first.", schemaPath))
	}

	// Load schema
	schema, err := c.schemaRepo.Load(ctx, schemaPath)
	if err != nil {
		c.logger.Error("Failed to load schema", "path", schemaPath, "error", err)
		return errors.WrapError(err, errors.SchemaErrorType, errors.SchemaInvalidCode, "Failed to load schema")
	}
	fmt.Printf("DEBUG: Loaded schema with %d files\n", len(schema.Files))
	for path := range schema.Files {
		fmt.Printf("  - %s\n", path)
	}

	// Generate data from schema
	c.logger.Info("Calling GenerateFromSchema", "folder", c.flags.FolderPath, "files", len(schema.Files))
	fmt.Printf("DEBUG: About to call GenerateFromSchema with folder=%s, schema files=%d\n", c.flags.FolderPath, len(schema.Files))
	
	outputData, err := c.dataService.GenerateFromSchema(ctx, schema, c.flags.FolderPath)
	if err != nil {
		c.logger.Error("Failed to generate data", "error", err)
		return err
	}
	
	fmt.Printf("DEBUG: GenerateFromSchema returned %d classes, %d records\n", outputData.GetClassCount(), outputData.GetTotalRecordCount())
	c.logger.Info("GenerateFromSchema completed", "classes", outputData.GetClassCount())

	// Determine output path
	outputPath := c.getDataOutputPath()

	// Ensure output directory exists
	if err := c.ensureOutputDirectory(outputPath); err != nil {
		return err
	}

	// Save output data
	fmt.Printf("DEBUG: Before SaveJSON - Classes: %d, Records: %d\n", outputData.GetClassCount(), outputData.GetTotalRecordCount())
	fmt.Printf("DEBUG: Data map has %d entries\n", len(outputData.Data))
	for className, records := range outputData.Data {
		fmt.Printf("  - %s: %d records\n", className, len(records))
	}
	
	if err := c.outputRepo.SaveJSON(ctx, outputData, outputPath); err != nil {
		c.logger.Error("Failed to save output data", "path", outputPath, "error", err)
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to save output data file")
	}

	// Success message
	fmt.Printf("Data generated successfully: %s\n", outputPath)
	fmt.Printf("Classes: %d\n", outputData.GetClassCount())
	fmt.Printf("Records: %d\n", outputData.GetTotalRecordCount())
	c.logger.Info("Data generation completed", 
		"path", outputPath, 
		"classes", outputData.GetClassCount(),
		"records", outputData.GetTotalRecordCount())

	return nil
}

// getSchemaPath determines the path to the schema file
func (c *DataCommand) getSchemaPath() string {
	const schemaFileName = "schema.yml"
	
	if c.flags.OutputPath == "" {
		return schemaFileName
	}
	return filepath.Join(c.flags.OutputPath, schemaFileName)
}

// getDataOutputPath determines the output path for the data file
func (c *DataCommand) getDataOutputPath() string {
	const dataFileName = "output.json"
	
	if c.flags.OutputPath == "" {
		return dataFileName
	}
	return filepath.Join(c.flags.OutputPath, dataFileName)
}

// ensureOutputDirectory ensures the output directory exists
func (c *DataCommand) ensureOutputDirectory(outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	if outputDir == "." {
		return nil // Current directory, no need to create
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		c.logger.Error("Failed to create output directory", "dir", outputDir, "error", err)
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to create output directory")
	}

	return nil
}