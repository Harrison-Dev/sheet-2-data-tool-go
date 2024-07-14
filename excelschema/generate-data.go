package excelschema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

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
			fmt.Printf("警告: 無法打開 Excel 文件 %s: %v\n", filePath, err)
			continue
		}

		for sheetName, sheetInfo := range fileInfo.Sheets {
			className := sheetInfo.ClassName
			rows, err := f.GetRows(sheetName)
			if err != nil {
				fmt.Printf("警告: 讀取 sheet %s 時發生錯誤: %v\n", sheetName, err)
				continue
			}

			if len(rows) >= sheetInfo.OffsetHeader {
				// 檢查是否存在 int 類型的 id 欄位
				idFieldIndex := -1
				for i, dc := range sheetInfo.DataClass {
					if dc.Name == "Id" {
						idFieldIndex = i
						break
					}
				}
				if idFieldIndex == -1 {
					return nil, fmt.Errorf("錯誤: sheet %s 中沒有找到 int 類型的 id 欄位", sheetName)
				}

				// 生成 schema 信息
				fields := make([]FieldInfo, len(sheetInfo.DataClass))
				for i, dc := range sheetInfo.DataClass {
					fields[i] = FieldInfo{
						Name:     dc.Name,
						DataType: dc.DataType,
					}
				}
				output.Schema[className] = fields

				// 生成數據
				sheetData := make([]interface{}, 0)
				for _, row := range rows[sheetInfo.OffsetHeader:] {
					rowData := make(map[string]interface{})
					for i, value := range row {
						if i < len(sheetInfo.DataClass) {
							fieldInfo := sheetInfo.DataClass[i]
							convertedValue, err := convertValue(value, fieldInfo.DataType)
							if err != nil {
								fmt.Printf("警告: 轉換字段 '%s' 的值時發生錯誤: %v\n", fieldInfo.Name, err)
								rowData[fieldInfo.Name] = value // 使用原始字符串值
							} else {
								rowData[fieldInfo.Name] = convertedValue
							}
						}
					}
					sheetData = append(sheetData, rowData)
				}
				output.Data[className] = sheetData
			} else {
				fmt.Printf("警告: sheet %s 的行數小於指定的 offset\n", sheetName)
			}
		}

		f.Close()
	}

	return output, nil
}

func SaveJSONOutput(output *JSONOutput, filename string) error {
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("轉換數據為 JSON 時發生錯誤: %v", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("保存數據文件時發生錯誤: %v", err)
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
