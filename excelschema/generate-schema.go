package excelschema

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

func GenerateBasicSchemaFromFolder(folderPath string) (*SchemaInfo, error) {
	schema := &SchemaInfo{Files: make(map[string]ExcelFileInfo)}

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".xlsx") || strings.HasSuffix(info.Name(), ".xls")) {
			if strings.HasPrefix(info.Name(), "~$") {
				fmt.Printf("跳過臨時文件: %s\n", info.Name())
				return nil
			}

			relativePath, err := filepath.Rel(folderPath, path)
			if err != nil {
				return fmt.Errorf("計算相對路徑時發生錯誤: %v", err)
			}
			excelInfo, err := processExcelFileBasic(path)
			if err != nil {
				fmt.Printf("警告: 處理檔案 %s 時發生錯誤: %v\n", relativePath, err)
				return nil
			}
			schema.Files[relativePath] = excelInfo
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("掃描資料夾時發生錯誤: %v", err)
	}

	return schema, nil
}

func processExcelFileBasic(filePath string) (ExcelFileInfo, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return ExcelFileInfo{}, err
	}
	defer f.Close()

	excelInfo := ExcelFileInfo{Sheets: make(map[string]SheetInfo)}

	for _, sheetName := range f.GetSheetList() {
		excelInfo.Sheets[sheetName] = SheetInfo{
			OffsetHeader: 2,
			ClassName:    sheetName,
			SheetName:    sheetName,
		}
	}

	return excelInfo, nil
}
