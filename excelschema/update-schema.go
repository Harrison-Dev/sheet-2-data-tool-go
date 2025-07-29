package excelschema

import (
	"fmt"
	"path/filepath"

	"github.com/xuri/excelize/v2"
)

func UpdateSchemaFromFolder(schema *SchemaInfo, excelDir string) error {
	for filePath, fileInfo := range schema.Files {
		fullPath := filepath.Join(excelDir, filePath)
		f, err := excelize.OpenFile(fullPath)
		if err != nil {
			fmt.Printf("Warning: unable to open Excel file %s: %v\n", filePath, err)
			continue
		}

		for sheetName, sheetInfo := range fileInfo.Sheets {
			rows, err := f.GetRows(sheetName)
			if err != nil {
				fmt.Printf("Warning: error reading sheet %s: %v\n", sheetName, err)
				continue
			}

			if len(rows) >= sheetInfo.OffsetHeader {
				headerRow := rows[sheetInfo.OffsetHeader-1]

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
				fmt.Printf("Warning: sheet %s has fewer rows than specified offset\n", sheetName)
			}
		}

		schema.Files[filePath] = fileInfo
		f.Close()
	}

	fmt.Println("Schema has been updated. Please manually set or modify data_type in schema.yml file.")
	return nil
}
