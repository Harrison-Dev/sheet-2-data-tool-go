package main

import (
	"flag"
	"fmt"
	"os"

	"excel-schema-generator/excelschema"
)

const (
	schemaFileName = "schema.yml"
	dataFileName   = "output.json"
)

func main() {
	// Define CLI flags
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
	dataCmd := flag.NewFlagSet("data", flag.ExitOnError)

	// Folder path flag for all commands
	var folderPath string
	generateCmd.StringVar(&folderPath, "folder", "", "Path to the Excel files folder")
	updateCmd.StringVar(&folderPath, "folder", "", "Path to the Excel files folder")
	dataCmd.StringVar(&folderPath, "folder", "", "Path to the Excel files folder")

	// Check if any arguments were provided
	if len(os.Args) < 2 {
		// No arguments, run GUI mode
		runGUI()
		return
	}

	// Parse the command
	switch os.Args[1] {
	case "generate":
		generateCmd.Parse(os.Args[2:])
	case "update":
		updateCmd.Parse(os.Args[2:])
	case "data":
		dataCmd.Parse(os.Args[2:])
	default:
		fmt.Println("Expected 'generate', 'update' or 'data' subcommands")
		os.Exit(1)
	}

	// Check if folder path is provided
	if folderPath == "" {
		fmt.Println("Please provide a folder path using the -folder flag")
		os.Exit(1)
	}

	// Execute the appropriate command
	if generateCmd.Parsed() {
		generateBasicSchema(folderPath)
	} else if updateCmd.Parsed() {
		updateSchema(folderPath)
	} else if dataCmd.Parsed() {
		generateData(folderPath)
	}
}

func generateBasicSchema(folderPath string) {
	schema, err := excelschema.GenerateBasicSchemaFromFolder(folderPath)
	if err != nil {
		fmt.Printf("Error generating schema: %v\n", err)
		return
	}
	err = schema.SaveToFile(schemaFileName)
	if err != nil {
		fmt.Printf("Error saving schema: %v\n", err)
		return
	}
	fmt.Printf("%s has been generated successfully in the current working directory\n", schemaFileName)
}

func updateSchema(folderPath string) {
	schema, err := excelschema.LoadSchemaFromFile(schemaFileName)
	if err != nil {
		fmt.Printf("Error loading schema: %v\n", err)
		return
	}
	err = excelschema.UpdateSchemaFromFolder(schema, folderPath)
	if err != nil {
		fmt.Printf("Error updating schema: %v\n", err)
		return
	}
	err = schema.SaveToFile(schemaFileName)
	if err != nil {
		fmt.Printf("Error saving updated schema: %v\n", err)
		return
	}
	fmt.Printf("%s has been updated successfully in the current working directory\n", schemaFileName)
}

func generateData(folderPath string) {
	schema, err := excelschema.LoadSchemaFromFile(schemaFileName)
	if err != nil {
		fmt.Printf("Error loading schema: %v\n", err)
		return
	}
	output, err := excelschema.GenerateDataFromFolder(schema, folderPath)
	if err != nil {
		fmt.Printf("Error generating data: %v\n", err)
		return
	}
	err = excelschema.SaveJSONOutput(output, dataFileName)
	if err != nil {
		fmt.Printf("Error saving data: %v\n", err)
		return
	}
	fmt.Printf("%s has been generated successfully in the current working directory\n", dataFileName)
}
