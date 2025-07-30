package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"excel-schema-generator/excelschema"
	"excel-schema-generator/pkg/logger"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	fyneDialog "fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/sqweek/dialog"
)

type GUI struct {
	window            fyne.Window
	config            *Config
	excelFolderEntry  *widget.Entry
	schemaFolderEntry *widget.Entry
	outputFolderEntry *widget.Entry
	statusLabel       *widget.Label
	linkLabel         *widget.Hyperlink
	content           *fyne.Container
}

func runGUI() {
	a := app.New()
	w := a.NewWindow("Excel Schema Generator")

	config, err := LoadConfig()
	if err != nil {
		fyneDialog.ShowError(err, w)
	}

	gui := &GUI{
		window: w,
		config: config,
	}

	gui.createUI()

	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}

func (g *GUI) createUI() {
	// Initialize logging
	logger.SetDefault(logger.New(logger.DefaultConfig()))
	
	g.excelFolderEntry = widget.NewEntry()
	g.excelFolderEntry.SetText(g.config.ExcelFolder)
	g.schemaFolderEntry = widget.NewEntry()
	g.schemaFolderEntry.SetText(g.config.SchemaFolder)
	g.outputFolderEntry = widget.NewEntry()
	g.outputFolderEntry.SetText(g.config.OutputFolder)
	g.statusLabel = widget.NewLabel("")
	g.linkLabel = widget.NewHyperlink("", nil)

	excelFolderBtn := widget.NewButton("Browse", func() {
		path, err := g.selectFolder("Select Excel Folder")
		if err != nil {
			fyneDialog.ShowError(err, g.window)
			return
		}
		g.excelFolderEntry.SetText(path)
		g.config.ExcelFolder = path
		g.saveConfig()
	})

	schemaFolderBtn := widget.NewButton("Browse", func() {
		path, err := g.selectFolder("Select Schema Folder")
		if err != nil {
			fyneDialog.ShowError(err, g.window)
			return
		}
		g.schemaFolderEntry.SetText(path)
		g.config.SchemaFolder = path
		g.saveConfig()
	})

	outputFolderBtn := widget.NewButton("Browse", func() {
		path, err := g.selectFolder("Select Output Folder")
		if err != nil {
			fyneDialog.ShowError(err, g.window)
			return
		}
		g.outputFolderEntry.SetText(path)
		g.config.OutputFolder = path
		g.saveConfig()
	})

	generateBtn := widget.NewButton("Generate Basic Schema", g.generateBasicSchema)
	updateBtn := widget.NewButton("Update Schema", g.updateSchema)
	dataBtn := widget.NewButton("Generate Data", g.generateData)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Excel Folder", Widget: container.NewBorder(nil, nil, nil, excelFolderBtn, g.excelFolderEntry)},
			{Text: "Schema Folder", Widget: container.NewBorder(nil, nil, nil, schemaFolderBtn, g.schemaFolderEntry)},
			{Text: "Output Folder", Widget: container.NewBorder(nil, nil, nil, outputFolderBtn, g.outputFolderEntry)},
		},
	}

	g.content = container.NewVBox(
		form,
		container.NewHBox(layout.NewSpacer(), generateBtn, updateBtn, dataBtn, layout.NewSpacer()),
		g.statusLabel,
		g.linkLabel,
	)

	g.window.SetContent(g.content)
}
func (g *GUI) selectFolder(title string) (string, error) {
	dir, err := dialog.Directory().Title(title).Browse()
	if err != nil {
		return "", err
	}
	return dir, nil
}

func (g *GUI) generateBasicSchema() {
	schema, err := excelschema.GenerateBasicSchemaFromFolder(g.config.ExcelFolder)
	if err != nil {
		g.showError("Error generating schema", err)
		return
	}
	schemaPath := g.config.GetSchemaPath()
	err = schema.SaveToFile(schemaPath)
	if err != nil {
		g.showError("Error saving schema", err)
		return
	}
	g.showSuccess(fmt.Sprintf("Schema generated: %s", schemaPath))
}

func (g *GUI) updateSchema() {
	schemaPath := g.config.GetSchemaPath()
	schema, err := excelschema.LoadSchemaFromFile(schemaPath)
	if err != nil {
		g.showError("Error loading schema", err)
		return
	}
	err = excelschema.UpdateSchemaFromFolder(schema, g.config.ExcelFolder)
	if err != nil {
		g.showError("Error updating schema", err)
		return
	}
	err = schema.SaveToFile(schemaPath)
	if err != nil {
		g.showError("Error saving updated schema", err)
		return
	}
	g.showSuccess(fmt.Sprintf("Schema updated: %s", schemaPath))
}

func (g *GUI) generateData() {
	schemaPath := g.config.GetSchemaPath()
	schema, err := excelschema.LoadSchemaFromFile(schemaPath)
	if err != nil {
		g.showError("Error loading schema", err)
		return
	}
	output, err := excelschema.GenerateDataFromFolder(schema, g.config.ExcelFolder)
	if err != nil {
		g.showError("Error generating data", err)
		return
	}
	outputPath := g.config.GetOutputPath()
	err = excelschema.SaveJSONOutput(output, outputPath)
	if err != nil {
		g.showError("Error saving data", err)
		return
	}
	g.showSuccess(fmt.Sprintf("Data generated: %s", outputPath))
}

func (g *GUI) showError(message string, err error) {
	fyneDialog.ShowError(fmt.Errorf("%s: %v", message, err), g.window)
	g.statusLabel.SetText(fmt.Sprintf("Error: %s", message))
}

func (g *GUI) showSuccess(message string) {
	fyneDialog.ShowInformation("Success", message, g.window)

	g.statusLabel.SetText(message)

	// Check if message contains file path
	if filepath.Ext(message) == ".yml" || filepath.Ext(message) == ".json" {
		// Extract file path
		path := message[strings.LastIndex(message, ":")+2:]

		// Update hyperlink
		g.linkLabel.SetText("Open containing folder")
		g.linkLabel.OnTapped = func() {
			dir := filepath.Dir(path)
			err := g.openFolder(dir)
			if err != nil {
				g.showError("Error opening folder", err)
			}
		}
	} else {
		// If no file path in message, hide link
		g.linkLabel.SetText("")
		g.linkLabel.OnTapped = nil
	}
}

func (g *GUI) openFolder(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default: // Assume Linux
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

func (g *GUI) saveConfig() {
	err := SaveConfig(g.config)
	if err != nil {
		g.showError("Error saving config", err)
	}
}
