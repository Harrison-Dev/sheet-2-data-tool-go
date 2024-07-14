package main

import (
	"fmt"

	"excel-schema-generator/excelschema"
)

func main() {
	fmt.Println("Excel Schema Tools")
	fmt.Println("1. 生成 Schema")
	fmt.Println("2. 讀取字段並更新 Schema 和數據")
	fmt.Print("請選擇功能 (1/2): ")

	var choice int
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		schema, err := excelschema.GenerateSchema()
		if err != nil {
			fmt.Printf("生成 Schema 時發生錯誤: %v\n", err)
			return
		}
		err = schema.SaveToFile("schema.yaml")
		if err != nil {
			fmt.Printf("保存 Schema 時發生錯誤: %v\n", err)
			return
		}
		fmt.Println("schema.yaml 已成功生成")

	case 2:
		schema, err := excelschema.LoadSchemaFromFile("schema.yaml")
		if err != nil {
			fmt.Printf("讀取 Schema 時發生錯誤: %v\n", err)
			return
		}

		err = excelschema.ReadFields(schema)
		if err != nil {
			fmt.Printf("讀取字段時發生錯誤: %v\n", err)
			return
		}

		err = schema.SaveToFile("schema.yaml")
		if err != nil {
			fmt.Printf("更新 Schema 時發生錯誤: %v\n", err)
			return
		}
		fmt.Println("schema.yaml 已成功更新，包含了 data class 信息")

		dataSchema, err := excelschema.GenerateDataSchema(schema)
		if err != nil {
			fmt.Printf("生成數據 Schema 時發生錯誤: %v\n", err)
			return
		}

		err = dataSchema.SaveToFile("data.yml")
		if err != nil {
			fmt.Printf("保存數據文件時發生錯誤: %v\n", err)
			return
		}
		fmt.Println("data.yml 已成功生成，包含了實際數據")

	default:
		fmt.Println("無效的選擇")
	}
}
