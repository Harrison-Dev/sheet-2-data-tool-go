package main

import (
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"excel-schema-generator/excelschema"
)

var selectedFolder string

func openGUI() {
	a := app.New()
	w := a.NewWindow("Excel Schema Generator")

	folderLabel := widget.NewLabel("No folder selected")
	selectFolderBtn := widget.NewButton("Select Excel Folder", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return
			}
			selectedFolder = uri.Path()
			folderLabel.SetText(selectedFolder)
		}, w)
	})

	generateBasicSchemaBtn := widget.NewButton("Generate Basic Schema", func() {
		if selectedFolder == "" {
			dialog.ShowInformation("Error", "Please select a folder first", w)
			return
		}
		schema, err := excelschema.GenerateBasicSchemaFromFolder(selectedFolder)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		err = schema.SaveToFile(filepath.Join(selectedFolder, "schema.yml"))
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("Success", "schema.yml has been generated", w)
	})

	updateSchemaBtn := widget.NewButton("Update Schema", func() {
		if selectedFolder == "" {
			dialog.ShowInformation("Error", "Please select a folder first", w)
			return
		}
		schema, err := excelschema.LoadSchemaFromFile(filepath.Join(selectedFolder, "schema.yml"))
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		err = excelschema.UpdateSchemaFromFolder(schema, selectedFolder)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		err = schema.SaveToFile(filepath.Join(selectedFolder, "schema.yml"))
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("Success", "schema.yml has been updated", w)
	})

	generateDataBtn := widget.NewButton("Generate Data", func() {
		if selectedFolder == "" {
			dialog.ShowInformation("Error", "Please select a folder first", w)
			return
		}
		schema, err := excelschema.LoadSchemaFromFile(filepath.Join(selectedFolder, "schema.yml"))
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		output, err := excelschema.GenerateDataFromFolder(schema, selectedFolder)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		err = excelschema.SaveJSONOutput(output, filepath.Join(selectedFolder, "output.json"))
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		dialog.ShowInformation("Success", "output.json has been generated", w)
	})

	content := container.NewVBox(
		folderLabel,
		selectFolderBtn,
		generateBasicSchemaBtn,
		updateSchemaBtn,
		generateDataBtn,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(300, 200))
	w.ShowAndRun()
}
