package filesystem

import (
	"context"
	"encoding/json"
	"io"

	"excel-schema-generator/internal/core/models"
	"excel-schema-generator/internal/ports"
	"excel-schema-generator/internal/utils/errors"
)

// OutputRepository implements the OutputRepository interface using filesystem
type OutputRepository struct {
	fileRepo ports.FileRepository
	logger   ports.LoggingService
}

// NewOutputRepository creates a new output repository
func NewOutputRepository(fileRepo ports.FileRepository, logger ports.LoggingService) *OutputRepository {
	return &OutputRepository{
		fileRepo: fileRepo,
		logger:   logger,
	}
}

// SaveJSON saves output data as JSON
func (r *OutputRepository) SaveJSON(ctx context.Context, output *models.OutputData, path string) error {
	r.logger.Debug("Saving output data as JSON", "path", path)

	if output == nil {
		return errors.NewValidationError(errors.ValidationRequiredFieldCode, "Output data cannot be nil")
	}

	// Marshal output to JSON with indentation
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return errors.WrapError(err, errors.InternalErrorType, errors.InternalStateInconsistentCode, "Failed to marshal output data to JSON")
	}

	// Write to file
	if err := r.fileRepo.Write(ctx, path, data); err != nil {
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to write output file")
	}

	r.logger.Info("Output data saved successfully", 
		"path", path, 
		"classes", output.GetClassCount(),
		"records", output.GetTotalRecordCount())
	return nil
}

// SaveWithWriter saves output data using a custom writer
func (r *OutputRepository) SaveWithWriter(ctx context.Context, output *models.OutputData, writer io.Writer) error {
	r.logger.Debug("Saving output data with custom writer")

	if output == nil {
		return errors.NewValidationError(errors.ValidationRequiredFieldCode, "Output data cannot be nil")
	}

	if writer == nil {
		return errors.NewValidationError(errors.ValidationRequiredFieldCode, "Writer cannot be nil")
	}

	// Create JSON encoder with indentation
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	// Encode output data
	if err := encoder.Encode(output); err != nil {
		return errors.WrapError(err, errors.InternalErrorType, errors.InternalStateInconsistentCode, "Failed to encode output data to JSON")
	}

	r.logger.Info("Output data saved with custom writer", 
		"classes", output.GetClassCount(),
		"records", output.GetTotalRecordCount())
	return nil
}

// LoadJSON loads output data from JSON
func (r *OutputRepository) LoadJSON(ctx context.Context, path string) (*models.OutputData, error) {
	r.logger.Debug("Loading output data from JSON", "path", path)

	// Read file content
	data, err := r.fileRepo.Read(ctx, path)
	if err != nil {
		return nil, errors.WrapError(err, errors.FileErrorType, errors.FileNotFoundCode, "Failed to read output file")
	}

	// Unmarshal JSON to output data
	var output models.OutputData
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, errors.WrapError(err, errors.InternalErrorType, errors.InternalStateInconsistentCode, "Failed to parse output JSON")
	}

	r.logger.Info("Output data loaded successfully", 
		"path", path, 
		"classes", output.GetClassCount(),
		"records", output.GetTotalRecordCount())
	return &output, nil
}