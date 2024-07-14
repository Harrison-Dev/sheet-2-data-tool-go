package excelschema

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sqweek/dialog"
	"github.com/xuri/excelize/v2"
)

func GenerateSchema() (*SchemaInfo, error) {
	folderPath, err := dialog.Directory().Title("請選擇要掃描的資料夾").Browse()
	if err != nil {
		return nil, fmt.Errorf("選擇資料夾時發生錯誤: %v", err)
	}

	if folderPath == "" {
		return nil, fmt.Errorf("沒有選擇資料夾")
	}

	schema := &SchemaInfo{Files: make(map[string]ExcelFileInfo)}

	err = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
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
			excelInfo, err := processExcelFile(path)
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

func processExcelFile(filePath string) (ExcelFileInfo, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return ExcelFileInfo{}, err
	}
	defer f.Close()

	excelInfo := ExcelFileInfo{Sheets: make(map[string]SheetInfo)}

	for _, sheetName := range f.GetSheetList() {
		excelInfo.Sheets[sheetName] = SheetInfo{
			OffsetHeader: 2,
			ClassName:    sheetName, // 設置默認的 class name 為 sheet name
			SheetName:    sheetName,
		}
	}

	return excelInfo, nil
}
