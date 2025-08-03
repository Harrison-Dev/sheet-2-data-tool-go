package data

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"excel-schema-generator/internal/core/models"
	"excel-schema-generator/internal/ports"
	"excel-schema-generator/internal/utils/errors"
)

// DataGenerator implements the DataService interface
type DataGenerator struct {
	excelRepo ports.ExcelRepository
	logger    ports.LoggingService
	validator ports.ValidationService
}

// NewDataGenerator creates a new data generator
func NewDataGenerator(
	excelRepo ports.ExcelRepository,
	logger ports.LoggingService,
	validator ports.ValidationService,
) *DataGenerator {
	return &DataGenerator{
		excelRepo: excelRepo,
		logger:    logger,
		validator: validator,
	}
}

// GenerateFromSchema generates JSON data from Excel files using a schema
func (g *DataGenerator) GenerateFromSchema(ctx context.Context, schema *models.SchemaInfo, folderPath string) (*models.OutputData, error) {
	fmt.Println("=== INSIDE GenerateFromSchema ===")
	fmt.Printf("Folder: %s, Schema files: %d\n", folderPath, len(schema.Files))
	g.logger.Info("Starting data generation from schema", "folder", folderPath, "files", len(schema.Files))

	// Create output data structure
	outputData := models.NewOutputData()
	outputData.Metadata.FileCount = len(schema.Files)
	
	totalRecords := 0

	// Process each file in the schema
	for relativePath, fileInfo := range schema.Files {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		fullPath := filepath.Join(folderPath, relativePath)
		fmt.Printf("Processing file: %s => %s\n", relativePath, fullPath)
		g.logger.Info("Processing file for data extraction", "file", relativePath, "path", fullPath)

		// Extract data from file
		classData, err := g.ExtractFromFile(ctx, fullPath, fileInfo)
		if err != nil {
			fmt.Printf("ERROR extracting from %s: %v\n", relativePath, err)
			g.logger.Error("Failed to extract data from file", "file", relativePath, "error", err)
			continue // Skip this file but continue with others
		}
		fmt.Printf("Extracted %d classes from %s\n", len(classData), relativePath)

		// Add extracted data to output
		for className, records := range classData {
			// Add data records
			outputData.AddData(className, records)
			totalRecords += len(records)
			
			// Add schema info if not already present
			if !outputData.HasClass(className) {
				// Extract field info from fileInfo for this class
				for sheetName, sheetInfo := range fileInfo.Sheets {
					if sheetInfo.ClassName == className || (sheetInfo.ClassName == "" && sheetName == className) {
						fields := make([]models.FieldInfo, 0, len(sheetInfo.DataClass))
						for _, field := range sheetInfo.DataClass {
							fields = append(fields, models.NewFieldInfo(field.Name, field.DataType))
						}
						outputData.AddSchema(className, fields)
						break
					}
				}
			}
		}
	}

	outputData.Metadata.RecordCount = totalRecords
	
	fmt.Printf("=== END OF GenerateFromSchema: totalRecords=%d ===\n", totalRecords)
	g.logger.Info("Data generation completed", 
		"files", len(schema.Files),
		"classes", outputData.GetClassCount(),
		"records", totalRecords)

	return outputData, nil
}

// ExtractFromFile extracts data from a single Excel file
func (g *DataGenerator) ExtractFromFile(ctx context.Context, filePath string, fileInfo models.ExcelFileInfo) (map[string][]interface{}, error) {
	g.logger.Info("Reading Excel file", "path", filePath)
	
	// Read Excel file
	excelData, err := g.excelRepo.Read(ctx, filePath)
	if err != nil {
		g.logger.Error("Failed to read Excel file", "path", filePath, "error", err)
		return nil, errors.WrapError(err, errors.ExcelErrorType, errors.ExcelInvalidFormatCode, "Failed to read Excel file")
	}
	
	g.logger.Info("Excel file read successfully", "sheets", len(excelData.Sheets))

	result := make(map[string][]interface{})

	// Process each sheet
	for sheetName, sheetInfo := range fileInfo.Sheets {
		_, exists := excelData.Sheets[sheetName]
		if !exists {
			g.logger.Warn("Sheet not found in Excel file", "sheet", sheetName, "file", filePath)
			continue
		}

		// Transform sheet data
		g.logger.Debug("Transforming sheet data", "sheet", sheetName, "rows", len(excelData.Sheets[sheetName].Rows))
		records, err := g.Transform(ctx, excelData, sheetInfo)
		if err != nil {
			g.logger.Error("Failed to transform sheet data", "sheet", sheetName, "error", err)
			continue
		}
		g.logger.Debug("Transformed records", "sheet", sheetName, "count", len(records))

		// Validate if needed
		if err := g.ValidateData(ctx, records, sheetInfo); err != nil {
			g.logger.Warn("Data validation failed", "sheet", sheetName, "error", err)
			// Continue anyway - validation errors shouldn't stop data extraction
		}

		// Use ClassName as the key for the output
		className := sheetInfo.ClassName
		if className == "" {
			className = sheetName
		}
		
		result[className] = records
	}

	return result, nil
}

// Transform transforms raw Excel data according to schema rules
func (g *DataGenerator) Transform(ctx context.Context, excelData *models.ExcelData, sheetInfo models.SheetInfo) ([]interface{}, error) {
	// For now, we'll just work with single sheet - this matches the interface requirement
	// In the actual implementation, we handle this at the ExtractFromFile level
	sheet, exists := excelData.Sheets[sheetInfo.SheetName]
	if !exists {
		return nil, errors.NewExcelError(errors.ExcelSheetNotFoundCode, fmt.Sprintf("Sheet '%s' not found", sheetInfo.SheetName))
	}
	records := make([]interface{}, 0, len(sheet.Rows))

	// Create field index map for quick lookup
	fieldIndexMap := make(map[string]int)
	for i, header := range sheet.Headers {
		fieldIndexMap[header] = i
	}

	// Process each row
	for rowIndex, row := range sheet.Rows {
		// Skip if row is empty
		if g.isEmptyRow(row) {
			continue
		}

		record := make(map[string]interface{})

		// Process each field defined in schema
		for _, field := range sheetInfo.DataClass {
			columnIndex, exists := fieldIndexMap[field.Name]
			if !exists {
				g.logger.Debug("Field not found in Excel headers", "field", field.Name)
				// Use default value if specified
				if field.Default != nil {
					record[field.Name] = field.Default
				}
				continue
			}

			// Get cell value
			var cellValue string
			if columnIndex < len(row) {
				cellValue = strings.TrimSpace(row[columnIndex])
			}

			// Convert value based on data type
			convertedValue, err := g.convertValue(cellValue, field.DataType)
			if err != nil {
				g.logger.Debug("Failed to convert value", 
					"field", field.Name, 
					"value", cellValue, 
					"type", field.DataType,
					"row", rowIndex+sheetInfo.OffsetHeader+1,
					"error", err)
				// Use default value or keep as string
				if field.Default != nil {
					record[field.Name] = field.Default
				} else {
					record[field.Name] = cellValue
				}
			} else {
				record[field.Name] = convertedValue
			}
		}

		records = append(records, record)
	}

	return records, nil
}

// ValidateData validates extracted data against schema rules
func (g *DataGenerator) ValidateData(ctx context.Context, data []interface{}, sheetInfo models.SheetInfo) error {
	// Basic validation - ensure required fields are present
	for _, record := range data {
		mapRecord, ok := record.(map[string]interface{})
		if !ok {
			continue
		}

		for _, field := range sheetInfo.DataClass {
			if field.Required {
				value, exists := mapRecord[field.Name]
				if !exists || value == nil || value == "" {
					return errors.NewValidationError(errors.ValidationRequiredFieldCode, 
						fmt.Sprintf("Required field '%s' is missing or empty", field.Name))
				}
			}
		}
	}

	return nil
}

// convertValue converts a string value to the specified data type
func (g *DataGenerator) convertValue(value string, dataType string) (interface{}, error) {
	// Handle empty values
	if value == "" {
		switch dataType {
		case "int", "float":
			return 0, nil
		case "bool":
			return false, nil
		default:
			return "", nil
		}
	}

	// Convert based on type
	switch dataType {
	case "int":
		return strconv.ParseInt(value, 10, 64)
	case "float":
		return strconv.ParseFloat(value, 64)
	case "bool":
		lowerValue := strings.ToLower(value)
		return lowerValue == "true" || lowerValue == "yes" || lowerValue == "1", nil
	default:
		return value, nil
	}
}

// isEmptyRow checks if all cells in a row are empty
func (g *DataGenerator) isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}