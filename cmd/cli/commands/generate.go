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

// GenerateCommand implements the generate schema command
type GenerateCommand struct {
	schemaService ports.SchemaService
	schemaRepo    ports.SchemaRepository
	logger        ports.LoggingService
	flags         flags.CommonFlags
}

// NewGenerateCommand creates a new generate command
func NewGenerateCommand(
	schemaService ports.SchemaService,
	schemaRepo ports.SchemaRepository,
	logger ports.LoggingService,
) *GenerateCommand {
	return &GenerateCommand{
		schemaService: schemaService,
		schemaRepo:    schemaRepo,
		logger:        logger,
	}
}

// Name returns the command name
func (c *GenerateCommand) Name() string {
	return "generate"
}

// Description returns the command description
func (c *GenerateCommand) Description() string {
	return "Generate a new schema from Excel files in a folder"
}

// SetupFlags sets up command-specific flags
func (c *GenerateCommand) SetupFlags(fs *flag.FlagSet) {
	flags.AddCommonFlags(fs, &c.flags)
}

// Execute executes the generate command
func (c *GenerateCommand) Execute(ctx context.Context, args []string) error {
	c.logger.Info("Starting schema generation command", 
		"folder", c.flags.FolderPath, 
		"output", c.flags.OutputPath)

	// Validate flags
	if err := c.flags.Validate(); err != nil {
		return errors.WrapError(err, errors.ValidationErrorType, errors.ValidationRequiredFieldCode, "Invalid command flags")
	}

	// Generate schema
	schema, err := c.schemaService.GenerateFromFolder(ctx, c.flags.FolderPath)
	if err != nil {
		c.logger.Error("Failed to generate schema", "error", err)
		return err
	}

	// Determine output path
	outputPath := c.getSchemaOutputPath()
	
	// Ensure output directory exists
	if err := c.ensureOutputDirectory(outputPath); err != nil {
		return err
	}

	// Save schema
	if err := c.schemaRepo.Save(ctx, schema, outputPath); err != nil {
		c.logger.Error("Failed to save schema", "path", outputPath, "error", err)
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to save schema file")
	}

	// Success message
	fmt.Printf("Schema generated successfully: %s\n", outputPath)
	c.logger.Info("Schema generation completed", "path", outputPath, "files", len(schema.Files))

	return nil
}

// getSchemaOutputPath determines the output path for the schema file
func (c *GenerateCommand) getSchemaOutputPath() string {
	const schemaFileName = "schema.yml"
	
	if c.flags.OutputPath == "" {
		return schemaFileName
	}
	return filepath.Join(c.flags.OutputPath, schemaFileName)
}

// ensureOutputDirectory ensures the output directory exists
func (c *GenerateCommand) ensureOutputDirectory(outputPath string) error {
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