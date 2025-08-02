package validation

import (
	"context"
	"fmt"

	"excel-schema-generator/internal/core/models"
	"excel-schema-generator/internal/ports"
	"excel-schema-generator/internal/utils/errors"
)

// ValidationService implements the ValidationService interface
type ValidationService struct {
	logger ports.LoggingService
}

// NewValidationService creates a new validation service
func NewValidationService(logger ports.LoggingService) *ValidationService {
	return &ValidationService{
		logger: logger,
	}
}

// ValidateExcelFile validates an Excel file structure
func (v *ValidationService) ValidateExcelFile(ctx context.Context, filePath string) error {
	v.logger.Debug("Validating Excel file", "path", filePath)

	if filePath == "" {
		return errors.NewValidationError(errors.ValidationRequiredFieldCode, "File path is required")
	}

	// Basic validation - in a real implementation, you'd do more thorough checks
	// such as checking file format, size limits, etc.
	
	return nil
}

// ValidateSchema validates a schema structure
func (v *ValidationService) ValidateSchema(ctx context.Context, schema *models.SchemaInfo) error {
	v.logger.Debug("Validating schema")

	if schema == nil {
		return errors.NewValidationError(errors.ValidationRequiredFieldCode, "Schema cannot be nil")
	}

	// Validate version
	if schema.Version == "" {
		return errors.NewSchemaError(errors.SchemaMissingFieldCode, "Schema version is required")
	}

	// Validate files
	if len(schema.Files) == 0 {
		return errors.NewSchemaError(errors.SchemaValidationFailedCode, "Schema must contain at least one file")
	}

	// Validate each file
	for relativePath, fileInfo := range schema.Files {
		if err := v.validateFileInfo(relativePath, fileInfo); err != nil {
			return err
		}
	}

	v.logger.Debug("Schema validation passed", "files", len(schema.Files))
	return nil
}

// ValidateDataTypes validates data types in extracted data
func (v *ValidationService) ValidateDataTypes(ctx context.Context, data []interface{}, fields []models.DataClassInfo) error {
	v.logger.Debug("Validating data types", "records", len(data), "fields", len(fields))

	if len(fields) == 0 {
		return errors.NewValidationError(errors.ValidationRequiredFieldCode, "Fields definition is required")
	}

	// Validate each record
	for i, record := range data {
		if err := v.validateRecord(record, fields, i); err != nil {
			return err
		}
	}

	v.logger.Debug("Data type validation passed")
	return nil
}

// ValidateRules validates custom validation rules
func (v *ValidationService) ValidateRules(ctx context.Context, data []interface{}, rules []models.ValidationRule) error {
	v.logger.Debug("Validating custom rules", "records", len(data), "rules", len(rules))

	// For now, just log that rules validation was requested
	// In a real implementation, you'd implement specific rule validation logic
	if len(rules) > 0 {
		v.logger.Info("Custom validation rules found but not yet implemented", "count", len(rules))
	}

	return nil
}

// validateFileInfo validates a single file info structure
func (v *ValidationService) validateFileInfo(relativePath string, fileInfo models.ExcelFileInfo) error {
	if fileInfo.FileName == "" {
		return errors.NewSchemaError(errors.SchemaMissingFieldCode, fmt.Sprintf("File name is required for file: %s", relativePath))
	}

	if len(fileInfo.Sheets) == 0 {
		return errors.NewSchemaError(errors.SchemaValidationFailedCode, fmt.Sprintf("File must contain at least one sheet: %s", relativePath))
	}

	// Validate each sheet
	for sheetName, sheetInfo := range fileInfo.Sheets {
		if err := v.validateSheetInfo(relativePath, sheetName, sheetInfo); err != nil {
			return err
		}
	}

	return nil
}

// validateSheetInfo validates a single sheet info structure
func (v *ValidationService) validateSheetInfo(relativePath, sheetName string, sheetInfo models.SheetInfo) error {
	if sheetInfo.SheetName == "" {
		return errors.NewSchemaError(errors.SchemaMissingFieldCode, fmt.Sprintf("Sheet name is required for sheet: %s in file: %s", sheetName, relativePath))
	}

	if sheetInfo.ClassName == "" {
		return errors.NewSchemaError(errors.SchemaMissingFieldCode, fmt.Sprintf("Class name is required for sheet: %s in file: %s", sheetName, relativePath))
	}

	if sheetInfo.OffsetHeader < 1 {
		return errors.NewSchemaError(errors.SchemaValidationFailedCode, fmt.Sprintf("Header offset must be at least 1 for sheet: %s in file: %s", sheetName, relativePath))
	}

	// Validate data class fields
	for i, dataClass := range sheetInfo.DataClass {
		if err := v.validateDataClass(relativePath, sheetName, i, dataClass); err != nil {
			return err
		}
	}

	return nil
}

// validateDataClass validates a single data class field
func (v *ValidationService) validateDataClass(relativePath, sheetName string, index int, dataClass models.DataClassInfo) error {
	if dataClass.Name == "" {
		return errors.NewSchemaError(errors.SchemaMissingFieldCode, fmt.Sprintf("Field name is required for field %d in sheet: %s, file: %s", index, sheetName, relativePath))
	}

	if dataClass.DataType == "" {
		return errors.NewSchemaError(errors.SchemaMissingFieldCode, fmt.Sprintf("Data type is required for field: %s in sheet: %s, file: %s", dataClass.Name, sheetName, relativePath))
	}

	// Validate data type is supported
	supportedTypes := map[string]bool{
		"string": true,
		"int":    true,
		"float":  true,
		"bool":   true,
	}

	if !supportedTypes[dataClass.DataType] {
		return errors.NewSchemaError(errors.SchemaValidationFailedCode, fmt.Sprintf("Unsupported data type '%s' for field: %s in sheet: %s, file: %s", dataClass.DataType, dataClass.Name, sheetName, relativePath))
	}

	return nil
}

// validateRecord validates a single data record against field definitions
func (v *ValidationService) validateRecord(record interface{}, fields []models.DataClassInfo, recordIndex int) error {
	recordMap, ok := record.(map[string]interface{})
	if !ok {
		return errors.NewValidationError(errors.ValidationInvalidTypeCode, fmt.Sprintf("Record %d is not a valid object", recordIndex))
	}

	// Check required fields
	for _, field := range fields {
		if field.Required {
			if _, exists := recordMap[field.Name]; !exists {
				return errors.NewValidationError(errors.ValidationRequiredFieldCode, fmt.Sprintf("Required field '%s' is missing in record %d", field.Name, recordIndex))
			}
		}
	}

	return nil
}