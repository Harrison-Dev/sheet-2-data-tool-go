package main

import (
	"encoding/json"
	"fmt"
	"os"

	"excel-schema-generator/excelschema"
)

func main() {
	fmt.Println("Excel Schema Tools")
	fmt.Println("1. 從 Excel 生成基本 Schema")
	fmt.Println("2. 更新 Schema (讀取 header)")
	fmt.Println("3. 根據 Schema 生成 Data (JSON 格式)")
	fmt.Print("請選擇功能 (1/2/3): ")

	var choice int
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		schema, err := excelschema.GenerateBasicSchema()
		if err != nil {
			fmt.Printf("生成基本 Schema 時發生錯誤: %v\n", err)
			return
		}
		err = schema.SaveToFile("schema.yml")
		if err != nil {
			fmt.Printf("保存 Schema 時發生錯誤: %v\n", err)
			return
		}
		fmt.Println("schema.yml 已成功生成")
		fmt.Println("請在 schema.yml 文件中手動設置 data_type")

	case 2:
		schema, err := excelschema.LoadSchemaFromFile("schema.yml")
		if err != nil {
			fmt.Printf("讀取 Schema 時發生錯誤: %v\n", err)
			return
		}

		err = excelschema.UpdateSchema(schema)
		if err != nil {
			fmt.Printf("更新 Schema 時發生錯誤: %v\n", err)
			return
		}

		err = schema.SaveToFile("schema.yml")
		if err != nil {
			fmt.Printf("保存更新後的 Schema 時發生錯誤: %v\n", err)
			return
		}
		fmt.Println("schema.yml 已成功更新")
		fmt.Println("請檢查並在需要時手動修改 schema.yml 中的 data_type")

	case 3:
		schema, err := excelschema.LoadSchemaFromFile("schema.yml")
		if err != nil {
			fmt.Printf("讀取 Schema 時發生錯誤: %v\n", err)
			return
		}

		output, err := excelschema.GenerateData(schema)
		if err != nil {
			fmt.Printf("生成數據時發生錯誤: %v\n", err)
			return
		}

		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Printf("轉換數據為 JSON 時發生錯誤: %v\n", err)
			return
		}

		err = os.WriteFile("output.json", jsonData, 0644)
		if err != nil {
			fmt.Printf("保存數據文件時發生錯誤: %v\n", err)
			return
		}
		fmt.Println("output.json 已成功生成，包含 schema 和 data 信息")

	default:
		fmt.Println("無效的選擇")
	}
}
