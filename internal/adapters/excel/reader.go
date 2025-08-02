package excel

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"excel-schema-generator/internal/core/models"
	"excel-schema-generator/internal/ports"
	"excel-schema-generator/internal/utils/errors"
	"github.com/xuri/excelize/v2"
)

// ExcelRepository implements the ExcelRepository interface
type ExcelRepository struct {
	logger ports.LoggingService
}

// NewExcelRepository creates a new Excel repository
func NewExcelRepository(logger ports.LoggingService) *ExcelRepository {
	return &ExcelRepository{
		logger: logger,
	}
}

// Read reads an Excel file and returns its data
func (r *ExcelRepository) Read(ctx context.Context, path string) (*models.ExcelData, error) {
	return r.ReadWithOptions(ctx, path, models.DefaultExcelProcessingOptions())
}

// ReadWithOptions reads an Excel file with specific options
func (r *ExcelRepository) ReadWithOptions(ctx context.Context, path string, options models.ExcelProcessingOptions) (*models.ExcelData, error) {
	r.logger.Debug("Reading Excel file", "path", path)

	// Validate file path
	if err := r.ValidateFile(ctx, path); err != nil {
		return nil, err
	}

	// Get file info
	fileInfo, err := r.GetFileInfo(ctx, path)
	if err != nil {
		return nil, errors.WrapError(err, errors.ExcelErrorType, errors.ExcelInvalidFormatCode, "Failed to get file info")
	}

	// Open Excel file
	f, err := excelize.OpenFile(path)
	if err != nil {
		r.logger.Error("Failed to open Excel file", "path", path, "error", err)
		return nil, r.handleExcelError(err, path)
	}
	defer f.Close()

	// Create Excel data container
	excelData := models.NewExcelData(*fileInfo)

	// Process each sheet
	sheetList := f.GetSheetList()
	r.logger.Debug("Found sheets", "count", len(sheetList), "sheets", sheetList)

	for _, sheetName := range sheetList {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		sheet, err := r.processSheet(f, sheetName, options)
		if err != nil {
			r.logger.Warn("Failed to process sheet", "sheet", sheetName, "error", err)
			// Continue with other sheets instead of failing completely
			continue
		}

		if !sheet.IsEmpty() || !options.SkipEmptyColumns {
			excelData.AddSheet(sheetName, sheet)
		}
	}

	r.logger.Info("Successfully read Excel file", "path", path, "sheets", len(excelData.Sheets))
	return &excelData, nil
}

// GetFileInfo retrieves metadata about an Excel file
func (r *ExcelRepository) GetFileInfo(ctx context.Context, path string) (*models.ExcelFile, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, errors.WrapError(err, errors.FileErrorType, errors.FileNotFoundCode, "File not found")
	}

	// Calculate checksum
	checksum, err := r.calculateChecksum(path)
	if err != nil {
		r.logger.Warn("Failed to calculate checksum", "path", path, "error", err)
		checksum = "" // Continue without checksum
	}

	fileInfo := models.NewExcelFile(
		path,
		filepath.Base(path),
		stat.Size(),
		stat.ModTime(),
	)
	fileInfo.Checksum = checksum

	return &fileInfo, nil
}

// ValidateFile validates that a file is a valid Excel file
func (r *ExcelRepository) ValidateFile(ctx context.Context, path string) error {
	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return errors.NewFileError(errors.FileNotFoundCode, fmt.Sprintf("File not found: %s", path))
		}
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Cannot access file")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".xlsx" && ext != ".xls" {
		return errors.NewExcelError(errors.ExcelInvalidFormatCode, fmt.Sprintf("Invalid file extension: %s. Expected .xlsx or .xls", ext))
	}

	// Check if it's a temporary file
	filename := filepath.Base(path)
	if strings.HasPrefix(filename, "~$") {
		return errors.NewExcelError(errors.ExcelInvalidFormatCode, "Temporary Excel file detected")
	}

	// Try to open the file to validate format
	f, err := excelize.OpenFile(path)
	if err != nil {
		return r.handleExcelError(err, path)
	}
	f.Close()

	return nil
}

// processSheet processes a single sheet from the Excel file
func (r *ExcelRepository) processSheet(f *excelize.File, sheetName string, options models.ExcelProcessingOptions) (models.ExcelSheet, error) {
	sheet := models.NewExcelSheet(sheetName)

	// Get all rows from the sheet
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return sheet, errors.WrapError(err, errors.ExcelErrorType, errors.ExcelSheetNotFoundCode, fmt.Sprintf("Failed to read sheet: %s", sheetName))
	}

	if len(rows) == 0 {
		return sheet, nil
	}

	// Process headers
	if options.HeaderRow > 0 && len(rows) >= options.HeaderRow {
		headerRow := rows[options.HeaderRow-1]
		if options.TrimWhitespace {
			headerRow = r.trimStringSlice(headerRow)
		}
		sheet.SetHeaders(headerRow)
	}

	// Process data rows
	startRow := options.HeaderRow
	if startRow <= 0 {
		startRow = 1
	}

	for i, row := range rows[startRow:] {
		// Check row limits
		if options.MaxRows > 0 && i >= options.MaxRows {
			break
		}

		// Skip empty rows if configured
		if options.SkipEmptyRows && r.isEmptyRow(row) {
			continue
		}

		// Trim whitespace if configured
		if options.TrimWhitespace {
			row = r.trimStringSlice(row)
		}

		// Limit columns if configured
		if options.MaxColumns > 0 && len(row) > options.MaxColumns {
			row = row[:options.MaxColumns]
		}

		sheet.AddRow(row)
	}

	return sheet, nil
}

// calculateChecksum calculates MD5 checksum of the file
func (r *ExcelRepository) calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// handleExcelError converts excelize errors to application errors
func (r *ExcelRepository) handleExcelError(err error, path string) error {
	errStr := err.Error()
	
	switch {
	case strings.Contains(errStr, "password"):
		return errors.NewExcelError(errors.ExcelPasswordProtectedCode, "Excel file is password protected")
	case strings.Contains(errStr, "not supported"):
		return errors.NewExcelError(errors.ExcelInvalidFormatCode, "Excel file format not supported")
	case strings.Contains(errStr, "corrupted") || strings.Contains(errStr, "invalid"):
		return errors.NewExcelError(errors.ExcelCorruptedCode, "Excel file appears to be corrupted")
	default:
		return errors.WrapError(err, errors.ExcelErrorType, errors.ExcelInvalidFormatCode, "Failed to process Excel file")
	}
}

// isEmptyRow checks if a row is empty (all cells are empty strings)
func (r *ExcelRepository) isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// trimStringSlice trims whitespace from all strings in a slice
func (r *ExcelRepository) trimStringSlice(slice []string) []string {
	result := make([]string, len(slice))
	for i, s := range slice {
		result[i] = strings.TrimSpace(s)
	}
	return result
}