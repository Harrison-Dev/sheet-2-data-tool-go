package filesystem

import (
	"context"

	"excel-schema-generator/internal/core/models"
	"excel-schema-generator/internal/ports"
	"excel-schema-generator/internal/utils/errors"
	"gopkg.in/yaml.v2"
)

// SchemaRepository implements the SchemaRepository interface using filesystem
type SchemaRepository struct {
	fileRepo ports.FileRepository
	logger   ports.LoggingService
}

// NewSchemaRepository creates a new schema repository
func NewSchemaRepository(fileRepo ports.FileRepository, logger ports.LoggingService) *SchemaRepository {
	return &SchemaRepository{
		fileRepo: fileRepo,
		logger:   logger,
	}
}

// Save saves a schema to storage
func (r *SchemaRepository) Save(ctx context.Context, schema *models.SchemaInfo, path string) error {
	r.logger.Debug("Saving schema", "path", path)

	if schema == nil {
		return errors.NewValidationError(errors.ValidationRequiredFieldCode, "Schema cannot be nil")
	}

	// Marshal schema to YAML
	data, err := yaml.Marshal(schema)
	if err != nil {
		return errors.WrapError(err, errors.SchemaErrorType, errors.SchemaInvalidCode, "Failed to marshal schema to YAML")
	}

	// Write to file
	if err := r.fileRepo.Write(ctx, path, data); err != nil {
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to write schema file")
	}

	r.logger.Info("Schema saved successfully", "path", path)
	return nil
}

// Load loads a schema from storage
func (r *SchemaRepository) Load(ctx context.Context, path string) (*models.SchemaInfo, error) {
	r.logger.Debug("Loading schema", "path", path)

	// Read file content
	data, err := r.fileRepo.Read(ctx, path)
	if err != nil {
		return nil, errors.WrapError(err, errors.FileErrorType, errors.FileNotFoundCode, "Failed to read schema file")
	}

	// Unmarshal YAML to schema
	var schema models.SchemaInfo
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, errors.WrapError(err, errors.SchemaErrorType, errors.SchemaInvalidCode, "Failed to parse schema YAML")
	}

	r.logger.Info("Schema loaded successfully", "path", path, "files", len(schema.Files))
	return &schema, nil
}

// Exists checks if a schema exists at the given path
func (r *SchemaRepository) Exists(ctx context.Context, path string) (bool, error) {
	return r.fileRepo.Exists(ctx, path)
}

// Delete removes a schema from storage
func (r *SchemaRepository) Delete(ctx context.Context, path string) error {
	r.logger.Debug("Deleting schema", "path", path)

	if err := r.fileRepo.Delete(ctx, path); err != nil {
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to delete schema file")
	}

	r.logger.Info("Schema deleted successfully", "path", path)
	return nil
}