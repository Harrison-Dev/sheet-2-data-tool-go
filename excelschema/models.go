package excelschema

import (
	"os"

	"gopkg.in/yaml.v2"
)

type DataClassInfo struct {
	Name     string `yaml:"name"`
	DataType string `yaml:"data_type"`
}

type SheetInfo struct {
	OffsetHeader int             `yaml:"offset_header"`
	ClassName    string          `yaml:"class_name"`
	SheetName    string          `yaml:"sheet_name"`
	DataClass    []DataClassInfo `yaml:"data_class,omitempty"`
	Data         [][]string      `yaml:"data,omitempty"`
}

type ExcelFileInfo struct {
	Sheets map[string]SheetInfo `yaml:"sheets"`
}

type SchemaInfo struct {
	Files map[string]ExcelFileInfo `yaml:"files"`
}

func (s *SchemaInfo) SaveToFile(filename string) error {
	yamlData, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, yamlData, 0644)
}

func LoadSchemaFromFile(filename string) (*SchemaInfo, error) {
	yamlData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var schema SchemaInfo
	err = yaml.Unmarshal(yamlData, &schema)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}
