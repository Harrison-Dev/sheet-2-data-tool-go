package excelschema

import (
	"path/filepath"

	"excel-schema-generator/pkg/logger"
	"github.com/xuri/excelize/v2"
)

func UpdateSchemaFromFolder(schema *SchemaInfo, excelDir string) error {
	for filePath, fileInfo := range schema.Files {
		fullPath := filepath.Join(excelDir, filePath)
		f, err := excelize.OpenFile(fullPath)
		if err != nil {
			logger.Warn("Unable to open Excel file", "file", filePath, "error", err)
			continue
		}

		for sheetName, sheetInfo := range fileInfo.Sheets {
			rows, err := f.GetRows(sheetName)
			if err != nil {
				logger.Warn("Error reading sheet", "sheet", sheetName, "file", filePath, "error", err)
				continue
			}

			if len(rows) >= sheetInfo.OffsetHeader {
				headerRow := rows[0] // 表頭永遠在第0行（第1行）

				// 保留現有的 DataClass 信息
				existingDataClass := make(map[string]DataClassInfo)
				for _, dc := range sheetInfo.DataClass {
					existingDataClass[dc.Name] = dc
				}

				sheetInfo.DataClass = make([]DataClassInfo, len(headerRow))

				for i, fieldName := range headerRow {
					if existing, ok := existingDataClass[fieldName]; ok {
						sheetInfo.DataClass[i] = existing
					} else {
						sheetInfo.DataClass[i] = DataClassInfo{
							Name:     fieldName,
							DataType: "string", // Set default data_type to string
						}
					}
				}

				fileInfo.Sheets[sheetName] = sheetInfo
			} else {
				logger.Warn("Sheet has insufficient rows", "sheet", sheetName, "file", filePath, "offset", sheetInfo.OffsetHeader, "rows", len(rows))
			}
		}

		schema.Files[filePath] = fileInfo
		f.Close()
	}

	logger.Info("Schema update completed", "message", "Please manually set or modify data_type in schema.yml file")
	return nil
}
