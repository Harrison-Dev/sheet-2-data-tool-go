package schema

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"excel-schema-generator/internal/core/models"
	"excel-schema-generator/internal/ports"
	"excel-schema-generator/internal/utils/errors"
)

// SchemaGenerator implements the SchemaService interface
type SchemaGenerator struct {
	excelRepo      ports.ExcelRepository
	fileRepo       ports.FileRepository
	logger         ports.LoggingService
	validator      ports.ValidationService
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator(
	excelRepo ports.ExcelRepository,
	fileRepo ports.FileRepository,
	logger ports.LoggingService,
	validator ports.ValidationService,
) *SchemaGenerator {
	return &SchemaGenerator{
		excelRepo: excelRepo,
		fileRepo:  fileRepo,
		logger:    logger,
		validator: validator,
	}
}

// GenerateFromFolder generates a new schema from Excel files in a folder
func (g *SchemaGenerator) GenerateFromFolder(ctx context.Context, folderPath string) (*models.SchemaInfo, error) {
	g.logger.Info("Starting schema generation", "folder", folderPath)

	// Validate folder path
	exists, err := g.fileRepo.Exists(ctx, folderPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewFileError(errors.DirectoryNotFoundCode, fmt.Sprintf("Folder not found: %s", folderPath))
	}

	// Get Excel files from folder
	excelFiles, err := g.getExcelFiles(ctx, folderPath)
	if err != nil {
		return nil, err
	}

	if len(excelFiles) == 0 {
		g.logger.Warn("No Excel files found in folder", "folder", folderPath)
		return nil, errors.NewValidationError(errors.ValidationRequiredFieldCode, "No Excel files found in the specified folder")
	}

	// Create new schema
	schema := models.NewSchemaInfo()
	schema.Metadata.Description = fmt.Sprintf("Generated schema from folder: %s", folderPath)

	// Process each Excel file
	for _, relativePath := range excelFiles {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		fullPath := filepath.Join(folderPath, relativePath)
		g.logger.Debug("Processing Excel file", "file", relativePath)

		fileInfo, err := g.processExcelFile(ctx, fullPath, relativePath)
		if err != nil {
			g.logger.Warn("Failed to process Excel file", "file", relativePath, "error", err)
			// Continue with other files instead of failing completely
			continue
		}

		schema.AddFile(relativePath, fileInfo)
	}

	// Validate generated schema
	if err := g.validator.ValidateSchema(ctx, schema); err != nil {
		return nil, errors.WrapError(err, errors.SchemaErrorType, errors.SchemaValidationFailedCode, "Generated schema is invalid")
	}

	g.logger.Info("Schema generation completed", "files", len(schema.Files), "sheets", schema.GetSheetCount())
	return schema, nil
}

// UpdateFromFolder updates an existing schema with Excel files from a folder
func (g *SchemaGenerator) UpdateFromFolder(ctx context.Context, schema *models.SchemaInfo, folderPath string) error {
	g.logger.Info("Starting schema update", "folder", folderPath)

	// Validate inputs
	if schema == nil {
		return errors.NewValidationError(errors.ValidationRequiredFieldCode, "Schema cannot be nil")
	}

	exists, err := g.fileRepo.Exists(ctx, folderPath)
	if err != nil {
		return err
	}
	if !exists {
		return errors.NewFileError(errors.DirectoryNotFoundCode, fmt.Sprintf("Folder not found: %s", folderPath))
	}

	// Get current Excel files
	excelFiles, err := g.getExcelFiles(ctx, folderPath)
	if err != nil {
		return err
	}

	// Track changes
	existingFiles := make(map[string]bool)
	for relativePath := range schema.Files {
		existingFiles[relativePath] = true
	}

	updatedCount := 0
	addedCount := 0

	// Process each current Excel file
	for _, relativePath := range excelFiles {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		fullPath := filepath.Join(folderPath, relativePath)
		g.logger.Debug("Processing Excel file for update", "file", relativePath)

		// Check if file needs update
		needsUpdate, err := g.checkFileNeedsUpdate(ctx, schema, relativePath, fullPath)
		if err != nil {
			g.logger.Warn("Failed to check if file needs update", "file", relativePath, "error", err)
			needsUpdate = true // Default to updating on error
		}

		if needsUpdate {
			// Get existing file info for merging
			existingFileInfo, _ := schema.GetFile(relativePath)
			
			// Process file with existing info for smart merge
			fileInfo, err := g.processExcelFileWithExisting(ctx, fullPath, relativePath, &existingFileInfo)
			if err != nil {
				g.logger.Warn("Failed to process Excel file during update", "file", relativePath, "error", err)
				continue
			}

			if existingFiles[relativePath] {
				updatedCount++
			} else {
				addedCount++
			}

			schema.AddFile(relativePath, fileInfo)
		}

		// Mark file as still existing
		delete(existingFiles, relativePath)
	}

	// Remove files that no longer exist
	removedCount := 0
	for relativePath := range existingFiles {
		schema.RemoveFile(relativePath)
		removedCount++
		g.logger.Debug("Removed missing file from schema", "file", relativePath)
	}

	// Update schema timestamp
	schema.UpdateTimestamp()

	// Validate updated schema
	if err := g.validator.ValidateSchema(ctx, schema); err != nil {
		return errors.WrapError(err, errors.SchemaErrorType, errors.SchemaValidationFailedCode, "Updated schema is invalid")
	}

	g.logger.Info("Schema update completed", 
		"added", addedCount, 
		"updated", updatedCount, 
		"removed", removedCount,
		"total_files", len(schema.Files))

	return nil
}

// Validate validates a schema for consistency and completeness
func (g *SchemaGenerator) Validate(ctx context.Context, schema *models.SchemaInfo) error {
	return g.validator.ValidateSchema(ctx, schema)
}

// Merge merges two schemas together
func (g *SchemaGenerator) Merge(ctx context.Context, base, additional *models.SchemaInfo) (*models.SchemaInfo, error) {
	if base == nil {
		return nil, errors.NewValidationError(errors.ValidationRequiredFieldCode, "Base schema cannot be nil")
	}
	if additional == nil {
		return nil, errors.NewValidationError(errors.ValidationRequiredFieldCode, "Additional schema cannot be nil")
	}

	g.logger.Debug("Merging schemas", "base_files", len(base.Files), "additional_files", len(additional.Files))

	// Create merged schema based on base
	merged := &models.SchemaInfo{
		Version:   base.Version,
		Metadata:  base.Metadata,
		Files:     make(map[string]models.ExcelFileInfo),
		CreatedAt: base.CreatedAt,
		UpdatedAt: time.Now(),
	}

	// Copy files from base
	for relativePath, fileInfo := range base.Files {
		merged.Files[relativePath] = fileInfo
	}

	// Add/update files from additional schema
	mergedCount := 0
	for relativePath, fileInfo := range additional.Files {
		if _, exists := merged.Files[relativePath]; exists {
			// File exists in both, merge sheets
			baseFileInfo := merged.Files[relativePath]
			mergedFileInfo := g.mergeFileInfo(baseFileInfo, fileInfo)
			merged.Files[relativePath] = mergedFileInfo
			mergedCount++
		} else {
			// New file, add it
			merged.Files[relativePath] = fileInfo
		}
	}

	// Update metadata
	merged.Metadata.Description = fmt.Sprintf("Merged schema - Base: %d files, Additional: %d files", len(base.Files), len(additional.Files))
	merged.UpdateTimestamp()

	// Validate merged schema
	if err := g.validator.ValidateSchema(ctx, merged); err != nil {
		return nil, errors.WrapError(err, errors.SchemaErrorType, errors.SchemaValidationFailedCode, "Merged schema is invalid")
	}

	g.logger.Info("Schema merge completed", "total_files", len(merged.Files), "merged_files", mergedCount)
	return merged, nil
}

// GetSchemaStatistics returns statistics about a schema
func (g *SchemaGenerator) GetSchemaStatistics(ctx context.Context, schema *models.SchemaInfo) (*ports.SchemaStatistics, error) {
	if schema == nil {
		return nil, errors.NewValidationError(errors.ValidationRequiredFieldCode, "Schema cannot be nil")
	}

	stats := &ports.SchemaStatistics{
		FileCount:   len(schema.Files),
		SheetCount:  schema.GetSheetCount(),
		LastUpdated: schema.UpdatedAt.Unix(),
	}

	// Calculate field count and total rows
	fieldCount := 0
	totalRows := 0
	for _, fileInfo := range schema.Files {
		for _, sheetInfo := range fileInfo.Sheets {
			fieldCount += len(sheetInfo.DataClass)
			totalRows += sheetInfo.RowCount
		}
	}

	stats.FieldCount = fieldCount
	stats.TotalRows = totalRows

	// Validate schema and collect any errors
	if err := g.validator.ValidateSchema(ctx, schema); err != nil {
		if appErr := errors.GetAppError(err); appErr != nil {
			stats.ValidationErrors = []string{appErr.Message}
		} else {
			stats.ValidationErrors = []string{err.Error()}
		}
	}

	return stats, nil
}

// processExcelFile processes a single Excel file and generates file info
func (g *SchemaGenerator) processExcelFile(ctx context.Context, fullPath, relativePath string) (models.ExcelFileInfo, error) {
	return g.processExcelFileWithExisting(ctx, fullPath, relativePath, nil)
}

// processExcelFileWithExisting processes a single Excel file with optional existing file info for merging
func (g *SchemaGenerator) processExcelFileWithExisting(ctx context.Context, fullPath, relativePath string, existingFileInfo *models.ExcelFileInfo) (models.ExcelFileInfo, error) {
	// Get Excel file metadata
	excelFile, err := g.excelRepo.GetFileInfo(ctx, fullPath)
	if err != nil {
		return models.ExcelFileInfo{}, err
	}

	// Read Excel data
	excelData, err := g.excelRepo.Read(ctx, fullPath)
	if err != nil {
		return models.ExcelFileInfo{}, err
	}

	// Create file info
	fileInfo := models.ExcelFileInfo{
		FileName:    excelFile.Name,
		FilePath:    relativePath,
		Checksum:    excelFile.Checksum,
		Sheets:      make(map[string]models.SheetInfo),
		LastUpdated: excelFile.LastModified,
	}

	// Process each sheet
	for sheetName, sheet := range excelData.Sheets {
		var existingSheetInfo *models.SheetInfo
		if existingFileInfo != nil {
			if existingSheet, exists := existingFileInfo.Sheets[sheetName]; exists {
				existingSheetInfo = &existingSheet
			}
		}
		
		sheetInfo := g.processSheetInfoWithExisting(sheetName, sheet, existingSheetInfo)
		fileInfo.Sheets[sheetName] = sheetInfo
	}

	return fileInfo, nil
}

// processSheetInfo processes sheet data and generates sheet info
func (g *SchemaGenerator) processSheetInfo(sheetName string, sheet models.ExcelSheet) models.SheetInfo {
	return g.processSheetInfoWithExisting(sheetName, sheet, nil)
}

// processSheetInfoWithExisting processes sheet data with optional existing sheet info for merging
func (g *SchemaGenerator) processSheetInfoWithExisting(sheetName string, sheet models.ExcelSheet, existingSheetInfo *models.SheetInfo) models.SheetInfo {
	sheetInfo := models.SheetInfo{
		SheetName:    sheetName,
		ClassName:    sheetName,
		OffsetHeader: 1, // Default header offset
		DataClass:    make([]models.DataClassInfo, 0),
		RowCount:     sheet.GetRowCount(),
	}

	// If we have existing sheet info, preserve manual settings
	if existingSheetInfo != nil {
		// Preserve manually configured values
		sheetInfo.ClassName = existingSheetInfo.ClassName
		sheetInfo.OffsetHeader = existingSheetInfo.OffsetHeader
		sheetInfo.ValidationRules = existingSheetInfo.ValidationRules
	}

	// Create map of existing fields for quick lookup
	existingFields := make(map[string]models.DataClassInfo)
	if existingSheetInfo != nil {
		for _, field := range existingSheetInfo.DataClass {
			existingFields[field.Name] = field
		}
	}

	// Generate data class info from headers
	for _, header := range sheet.Headers {
		if header != "" {
			dataClass := models.DataClassInfo{
				Name:     header,
				DataType: g.detectDataType(sheet, header),
				Required: false, // Default to not required
			}

			// If this field exists in the existing schema, preserve manual settings
			if existingField, exists := existingFields[header]; exists {
				// Preserve all manually configured values
				dataClass.Required = existingField.Required
				dataClass.Default = existingField.Default
				dataClass.Description = existingField.Description
				
				// Preserve existing DataType if it has been manually modified
				// We consider it manually modified if:
				// 1. The existing type is different from what auto-detection would give
				// 2. OR the existing type is not "string" (indicating manual configuration)
				autoDetectedType := g.detectDataType(sheet, header)
				if existingField.DataType != autoDetectedType || existingField.DataType != "string" {
					dataClass.DataType = existingField.DataType
				}
			}

			sheetInfo.DataClass = append(sheetInfo.DataClass, dataClass)
		}
	}

	return sheetInfo
}

// detectDataType attempts to detect the data type of a column
func (g *SchemaGenerator) detectDataType(sheet models.ExcelSheet, columnName string) string {
	// Find column index
	columnIndex := -1
	for i, header := range sheet.Headers {
		if header == columnName {
			columnIndex = i
			break
		}
	}

	if columnIndex == -1 {
		return "string" // Default type
	}

	// Sample first few rows to detect type
	sampleSize := 10
	if len(sheet.Rows) < sampleSize {
		sampleSize = len(sheet.Rows)
	}

	// Track type candidates
	hasInt := true
	hasFloat := true
	hasBool := true
	nonEmptyCount := 0

	// Check all sample values
	for i := 0; i < sampleSize; i++ {
		if i < len(sheet.Rows) && columnIndex < len(sheet.Rows[i]) {
			value := strings.TrimSpace(sheet.Rows[i][columnIndex])
			if value == "" {
				continue // Skip empty values
			}
			
			nonEmptyCount++
			
			// Check for boolean
			lowerValue := strings.ToLower(value)
			if hasBool && lowerValue != "true" && lowerValue != "false" && lowerValue != "yes" && lowerValue != "no" && lowerValue != "0" && lowerValue != "1" {
				hasBool = false
			}
			
			// Check for integer
			if hasInt {
				if _, err := strconv.ParseInt(value, 10, 64); err != nil {
					hasInt = false
				}
			}
			
			// Check for float
			if hasFloat {
				if _, err := strconv.ParseFloat(value, 64); err != nil {
					hasFloat = false
				}
			}
		}
	}

	// If no non-empty values found, default to string
	if nonEmptyCount == 0 {
		return "string"
	}

	// Determine type based on what's still valid
	// Priority: bool > int > float > string
	if hasBool {
		return "bool"
	}
	if hasInt {
		return "int"
	}
	if hasFloat {
		return "float"
	}
	
	return "string"
}

// getExcelFiles gets a list of Excel files from a folder
func (g *SchemaGenerator) getExcelFiles(ctx context.Context, folderPath string) ([]string, error) {
	files, err := g.fileRepo.List(ctx, folderPath, "")
	if err != nil {
		return nil, err
	}

	var excelFiles []string
	for _, file := range files {
		ext := filepath.Ext(file)
		if ext == ".xlsx" || ext == ".xls" {
			// Skip temporary files
			filename := filepath.Base(file)
			if !g.isTempFile(filename) {
				excelFiles = append(excelFiles, file)
			}
		}
	}

	return excelFiles, nil
}

// isTempFile checks if a filename represents a temporary Excel file
func (g *SchemaGenerator) isTempFile(filename string) bool {
	return len(filename) > 2 && filename[:2] == "~$"
}

// checkFileNeedsUpdate checks if a file needs to be updated in the schema
func (g *SchemaGenerator) checkFileNeedsUpdate(ctx context.Context, schema *models.SchemaInfo, relativePath, fullPath string) (bool, error) {
	// Check if file exists in schema
	existingFileInfo, exists := schema.GetFile(relativePath)
	if !exists {
		return true, nil // New file, needs to be added
	}

	// Get current file info
	currentFileInfo, err := g.excelRepo.GetFileInfo(ctx, fullPath)
	if err != nil {
		return true, err // Error getting file info, assume update needed
	}

	// Compare checksums if available
	if existingFileInfo.Checksum != "" && currentFileInfo.Checksum != "" {
		return existingFileInfo.Checksum != currentFileInfo.Checksum, nil
	}

	// Compare modification times
	return existingFileInfo.LastUpdated.Before(currentFileInfo.LastModified), nil
}

// mergeFileInfo merges two file info structures
func (g *SchemaGenerator) mergeFileInfo(base, additional models.ExcelFileInfo) models.ExcelFileInfo {
	merged := base

	// Use the more recent file info
	if additional.LastUpdated.After(base.LastUpdated) {
		merged.LastUpdated = additional.LastUpdated
		merged.Checksum = additional.Checksum
	}

	// Merge sheets
	for sheetName, additionalSheet := range additional.Sheets {
		if baseSheet, exists := merged.Sheets[sheetName]; exists {
			// Merge sheet info - prefer additional if it has more data
			if len(additionalSheet.DataClass) > len(baseSheet.DataClass) {
				merged.Sheets[sheetName] = additionalSheet
			}
		} else {
			// New sheet, add it
			merged.Sheets[sheetName] = additionalSheet
		}
	}

	return merged
}