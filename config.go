package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	ExcelFolder    string `json:"excelFolder"`
	SchemaFolder   string `json:"schemaFolder"`
	OutputFolder   string `json:"outputFolder"`
	SchemaFileName string `json:"schemaFileName"`
	OutputFileName string `json:"outputFileName"`
}

const configFileName = "config.json"

func LoadConfig() (*Config, error) {
	config := &Config{
		SchemaFileName: "schema.yml",
		OutputFileName: "output.json",
	}

	data, err := os.ReadFile(configFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}

	err = json.Unmarshal(data, config)
	return config, err
}

func SaveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFileName, data, 0644)
}

func (c *Config) GetSchemaPath() string {
	return filepath.Join(c.SchemaFolder, c.SchemaFileName)
}

func (c *Config) GetOutputPath() string {
	return filepath.Join(c.OutputFolder, c.OutputFileName)
}
