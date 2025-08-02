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

// UpdateCommand implements the update schema command
type UpdateCommand struct {
	schemaService ports.SchemaService
	schemaRepo    ports.SchemaRepository
	logger        ports.LoggingService
	flags         flags.CommonFlags
}

// NewUpdateCommand creates a new update command
func NewUpdateCommand(
	schemaService ports.SchemaService,
	schemaRepo ports.SchemaRepository,
	logger ports.LoggingService,
) *UpdateCommand {
	return &UpdateCommand{
		schemaService: schemaService,
		schemaRepo:    schemaRepo,
		logger:        logger,
	}
}

// Name returns the command name
func (c *UpdateCommand) Name() string {
	return "update"
}

// Description returns the command description
func (c *UpdateCommand) Description() string {
	return "Update an existing schema with Excel files from a folder"
}

// SetupFlags sets up command-specific flags
func (c *UpdateCommand) SetupFlags(fs *flag.FlagSet) {
	flags.AddCommonFlags(fs, &c.flags)
}

// Execute executes the update command
func (c *UpdateCommand) Execute(ctx context.Context, args []string) error {
	c.logger.Info("Starting schema update command", 
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
		return errors.NewSchemaError(errors.FileNotFoundCode, fmt.Sprintf("Schema file not found: %s. Use 'generate' command to create a new schema.", schemaPath))
	}

	// Load existing schema
	schema, err := c.schemaRepo.Load(ctx, schemaPath)
	if err != nil {
		c.logger.Error("Failed to load schema", "path", schemaPath, "error", err)
		return errors.WrapError(err, errors.SchemaErrorType, errors.SchemaInvalidCode, "Failed to load existing schema")
	}

	// Update schema with new data
	if err := c.schemaService.UpdateFromFolder(ctx, schema, c.flags.FolderPath); err != nil {
		c.logger.Error("Failed to update schema", "error", err)
		return err
	}

	// Ensure output directory exists
	if err := c.ensureOutputDirectory(schemaPath); err != nil {
		return err
	}

	// Save updated schema
	if err := c.schemaRepo.Save(ctx, schema, schemaPath); err != nil {
		c.logger.Error("Failed to save updated schema", "path", schemaPath, "error", err)
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to save updated schema file")
	}

	// Success message
	fmt.Printf("Schema updated successfully: %s\n", schemaPath)
	c.logger.Info("Schema update completed", "path", schemaPath, "files", len(schema.Files))

	return nil
}

// getSchemaPath determines the path to the schema file
func (c *UpdateCommand) getSchemaPath() string {
	const schemaFileName = "schema.yml"
	
	if c.flags.OutputPath == "" {
		return schemaFileName
	}
	return filepath.Join(c.flags.OutputPath, schemaFileName)
}

// ensureOutputDirectory ensures the output directory exists
func (c *UpdateCommand) ensureOutputDirectory(outputPath string) error {
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