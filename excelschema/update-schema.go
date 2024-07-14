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
			fmt.Printf("警告: 無法打開 Excel 文件 %s: %v\n", filePath, err)
			continue
		}

		for sheetName, sheetInfo := range fileInfo.Sheets {
			rows, err := f.GetRows(sheetName)
			if err != nil {
				fmt.Printf("警告: 讀取 sheet %s 時發生錯誤: %v\n", sheetName, err)
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
							DataType: "string", // 設置默認 data_type 為 string
						}
					}
				}

				fileInfo.Sheets[sheetName] = sheetInfo
			} else {
				fmt.Printf("警告: sheet %s 的行數小於指定的 offset\n", sheetName)
			}
		}

		schema.Files[filePath] = fileInfo
		f.Close()
	}

	fmt.Println("Schema 已更新。請在 schema.yml 文件中手動設置或修改 data_type。")
	return nil
}
