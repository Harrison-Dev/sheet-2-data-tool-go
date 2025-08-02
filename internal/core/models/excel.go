package models

import (
	"time"
)

// ExcelFile represents a physical Excel file with its metadata
type ExcelFile struct {
	Path         string
	Name         string
	Size         int64
	LastModified time.Time
	Checksum     string
}

// ExcelSheet represents a sheet within an Excel file
type ExcelSheet struct {
	Name    string
	Rows    [][]string
	Headers []string
}

// ExcelData represents the complete data extracted from an Excel file
type ExcelData struct {
	File   ExcelFile
	Sheets map[string]ExcelSheet
}

// CellValue represents a single cell value with type information
type CellValue struct {
	Value    interface{}
	DataType string
	Formula  string
}

// ExcelRow represents a row of data with cell information
type ExcelRow struct {
	Index int
	Cells map[string]CellValue
}

// ExcelProcessingOptions defines options for Excel file processing
type ExcelProcessingOptions struct {
	SkipEmptyRows    bool
	SkipEmptyColumns bool
	MaxRows          int
	MaxColumns       int
	HeaderRow        int
	TrimWhitespace   bool
}

// DefaultExcelProcessingOptions returns default processing options
func DefaultExcelProcessingOptions() ExcelProcessingOptions {
	return ExcelProcessingOptions{
		SkipEmptyRows:    true,
		SkipEmptyColumns: true,
		MaxRows:          10000,
		MaxColumns:       100,
		HeaderRow:        1,
		TrimWhitespace:   true,
	}
}

// NewExcelFile creates a new ExcelFile instance
func NewExcelFile(path, name string, size int64, lastModified time.Time) ExcelFile {
	return ExcelFile{
		Path:         path,
		Name:         name,
		Size:         size,
		LastModified: lastModified,
	}
}

// NewExcelSheet creates a new ExcelSheet instance
func NewExcelSheet(name string) ExcelSheet {
	return ExcelSheet{
		Name:   name,
		Rows:   make([][]string, 0),
		Headers: make([]string, 0),
	}
}

// AddRow adds a row to the Excel sheet
func (s *ExcelSheet) AddRow(row []string) {
	s.Rows = append(s.Rows, row)
}

// SetHeaders sets the headers for the Excel sheet
func (s *ExcelSheet) SetHeaders(headers []string) {
	s.Headers = headers
}

// GetRowCount returns the number of rows in the sheet
func (s *ExcelSheet) GetRowCount() int {
	return len(s.Rows)
}

// GetColumnCount returns the number of columns based on headers
func (s *ExcelSheet) GetColumnCount() int {
	return len(s.Headers)
}

// IsEmpty returns true if the sheet has no data rows
func (s *ExcelSheet) IsEmpty() bool {
	return len(s.Rows) == 0
}

// NewExcelData creates a new ExcelData instance
func NewExcelData(file ExcelFile) ExcelData {
	return ExcelData{
		File:   file,
		Sheets: make(map[string]ExcelSheet),
	}
}

// AddSheet adds a sheet to the Excel data
func (e *ExcelData) AddSheet(name string, sheet ExcelSheet) {
	e.Sheets[name] = sheet
}

// GetSheet retrieves a sheet by name
func (e *ExcelData) GetSheet(name string) (ExcelSheet, bool) {
	sheet, exists := e.Sheets[name]
	return sheet, exists
}

// GetSheetNames returns all sheet names
func (e *ExcelData) GetSheetNames() []string {
	names := make([]string, 0, len(e.Sheets))
	for name := range e.Sheets {
		names = append(names, name)
	}
	return names
}

// GetSheetCount returns the number of sheets
func (e *ExcelData) GetSheetCount() int {
	return len(e.Sheets)
}

// GetTotalRowCount returns the total number of rows across all sheets
func (e *ExcelData) GetTotalRowCount() int {
	total := 0
	for _, sheet := range e.Sheets {
		total += sheet.GetRowCount()
	}
	return total
}