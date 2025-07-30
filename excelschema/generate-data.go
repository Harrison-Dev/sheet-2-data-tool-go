package excelschema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"excel-schema-generator/pkg/logger"
	"github.com/xuri/excelize/v2"
)

type JSONOutput struct {
	Schema map[string][]FieldInfo   `json:"schema"`
	Data   map[string][]interface{} `json:"data"`
}

type FieldInfo struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
}

func GenerateDataFromFolder(schema *SchemaInfo, excelDir string) (*JSONOutput, error) {
	output := &JSONOutput{
		Schema: make(map[string][]FieldInfo),
		Data:   make(map[string][]interface{}),
	}

	for filePath, fileInfo := range schema.Files {
		fullPath := filepath.Join(excelDir, filePath)
		f, err := excelize.OpenFile(fullPath)
		if err != nil {
			logger.Warn("Unable to open Excel file", "file", filePath, "error", err)
			continue
		}

		for sheetName, sheetInfo := range fileInfo.Sheets {
			className := sheetInfo.ClassName
			rows, err := f.GetRows(sheetName)
			if err != nil {
				logger.Warn("Error reading sheet", "sheet", sheetName, "file", filePath, "error", err)
				continue
			}

			if len(rows) >= sheetInfo.OffsetHeader {
				// Check if there's an Id field
				hasIdField := false
				for _, dc := range sheetInfo.DataClass {
					if dc.Name == "Id" {
						hasIdField = true
						break
					}
				}

				// Generate schema information
				var fields []FieldInfo
				if !hasIdField {
					// Auto-generate Id field if not present
					logger.Info("No Id field found, auto-generating Id field", "sheet", sheetName, "file", filePath)
					fields = make([]FieldInfo, len(sheetInfo.DataClass)+1)
					fields[0] = FieldInfo{
						Name:     "Id",
						DataType: "int",
					}
					for i, dc := range sheetInfo.DataClass {
						fields[i+1] = FieldInfo{
							Name:     dc.Name,
							DataType: dc.DataType,
						}
					}
				} else {
					fields = make([]FieldInfo, len(sheetInfo.DataClass))
					for i, dc := range sheetInfo.DataClass {
						fields[i] = FieldInfo{
							Name:     dc.Name,
							DataType: dc.DataType,
						}
					}
				}
				output.Schema[className] = fields

				// Generate data
				sheetData := make([]interface{}, 0)
				for rowIndex, row := range rows[sheetInfo.OffsetHeader:] {
					rowData := make(map[string]interface{})
					
					if !hasIdField {
						// Add auto-generated Id starting from 0
						rowData["Id"] = rowIndex
					}
					
					for i, value := range row {
						if i < len(sheetInfo.DataClass) {
							fieldInfo := sheetInfo.DataClass[i]
							convertedValue, err := convertValue(value, fieldInfo.DataType)
							if err != nil {
								logger.Warn("Error converting field value", "field", fieldInfo.Name, "value", value, "type", fieldInfo.DataType, "error", err)
								rowData[fieldInfo.Name] = value // Use original string value
							} else {
								rowData[fieldInfo.Name] = convertedValue
							}
						}
					}
					sheetData = append(sheetData, rowData)
				}
				output.Data[className] = sheetData
			} else {
				logger.Warn("Sheet has insufficient rows", "sheet", sheetName, "file", filePath, "offset", sheetInfo.OffsetHeader, "rows", len(rows))
			}
		}

		f.Close()
	}

	return output, nil
}

func SaveJSONOutput(output *JSONOutput, filename string) error {
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		logger.Error("Failed to convert data to JSON", "error", err)
		return fmt.Errorf("error converting data to JSON: %v", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		logger.Error("Failed to save data file", "file", filename, "error", err)
		return fmt.Errorf("error saving data file: %v", err)
	}

	return nil
}
func convertValue(value string, dataType string) (interface{}, error) {
	switch dataType {
	case "string":
		return value, nil
	case "int":
		return strconv.Atoi(value)
	case "float":
		return strconv.ParseFloat(value, 64)
	case "bool":
		return strconv.ParseBool(value)
	default:
		return value, nil
	}
}
