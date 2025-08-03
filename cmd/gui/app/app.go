package app

import (
	"context"
	"fmt"

	"excel-schema-generator/gdrive"
	"excel-schema-generator/internal/adapters/filesystem"
	"excel-schema-generator/internal/core/models"
	"excel-schema-generator/internal/core/schema"
	"excel-schema-generator/internal/utils/errors"
	"excel-schema-generator/pkg/logger"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// GUIApp represents the GUI application
type GUIApp struct {
	app             fyne.App
	window          fyne.Window
	name            string
	version         string
	logger          *logger.Logger
	schemaGenerator *schema.SchemaGenerator
	fileRepo        *filesystem.FileRepository
	errorHandler    *errors.ErrorHandler
	
	// UI components
	excelFolderEntry  *widget.Entry
	schemaFolderEntry *widget.Entry
	outputFolderEntry *widget.Entry
	statusLabel       *widget.Label
	progressBar       *widget.ProgressBar
	
	// Google Drive download components
	credentialsEntry *widget.Entry
	driveLinkEntry   *widget.Entry
	downloadOutputEntry *widget.Entry
}

// NewGUIApp creates a new GUI application
func NewGUIApp(name, version string, logger *logger.Logger) *GUIApp {
	fyneApp := app.New()
	window := fyneApp.NewWindow(fmt.Sprintf("%s v%s", name, version))
	
	return &GUIApp{
		app:     fyneApp,
		window:  window,
		name:    name,
		version: version,
		logger:  logger,
	}
}

// SetDependencies sets the application dependencies
func (a *GUIApp) SetDependencies(
	schemaGenerator *schema.SchemaGenerator,
	fileRepo *filesystem.FileRepository,
	errorHandler *errors.ErrorHandler,
) {
	a.schemaGenerator = schemaGenerator
	a.fileRepo = fileRepo
	a.errorHandler = errorHandler
}

// Run runs the GUI application
func (a *GUIApp) Run() error {
	a.logger.Info("Starting GUI application", "name", a.name, "version", a.version)
	
	// Set app icon and theme
	a.app.SetIcon(theme.DocumentIcon())
	
	// Setup UI
	a.setupUI()
	
	// Configure window
	a.window.Resize(fyne.NewSize(900, 700))
	a.window.CenterOnScreen()
	a.window.SetFixedSize(false)
	
	// Show and run
	a.window.ShowAndRun()
	
	return nil
}

// setupUI sets up the user interface
func (a *GUIApp) setupUI() {
	// Initialize UI components
	a.initializeComponents()
	
	// Create layout
	content := a.createLayout()
	
	// Set window content
	a.window.SetContent(content)
}

// initializeComponents initializes UI components
func (a *GUIApp) initializeComponents() {
	// Entry fields with better styling
	a.excelFolderEntry = widget.NewEntry()
	a.excelFolderEntry.SetPlaceHolder("Select folder containing Excel files...")
	a.excelFolderEntry.Disable() // Read-only, use browse button
	
	a.schemaFolderEntry = widget.NewEntry()
	a.schemaFolderEntry.SetPlaceHolder("Select folder for schema files...")
	a.schemaFolderEntry.Disable() // Read-only, use browse button
	
	a.outputFolderEntry = widget.NewEntry()
	a.outputFolderEntry.SetPlaceHolder("Select output folder...")
	a.outputFolderEntry.Disable() // Read-only, use browse button
	
	// Google Drive download components
	a.credentialsEntry = widget.NewEntry()
	a.credentialsEntry.SetPlaceHolder("Select Google credentials JSON file...")
	a.credentialsEntry.Disable() // Read-only, use browse button
	
	a.driveLinkEntry = widget.NewEntry()
	a.driveLinkEntry.SetPlaceHolder("Enter Google Drive folder link...")
	
	a.downloadOutputEntry = widget.NewEntry()
	a.downloadOutputEntry.SetPlaceHolder("Select download output folder...")
	a.downloadOutputEntry.Disable() // Read-only, use browse button
	
	// Status components with better styling
	a.statusLabel = widget.NewLabel("Ready")
	a.statusLabel.Alignment = fyne.TextAlignCenter
	
	a.progressBar = widget.NewProgressBar()
	a.progressBar.Hide()
}

// createLayout creates the main application layout
func (a *GUIApp) createLayout() *fyne.Container {
	// Create header with title and logo
	header := a.createHeader()
	
	// Create tabs for different features
	tabs := container.NewAppTabs(
		container.NewTabItem("Schema Generation", a.createSchemaTab()),
		container.NewTabItem("Google Drive Download", a.createDriveDownloadTab()),
	)
	
	// Status section with enhanced feedback
	statusSection := a.createStatusSection()
	
	// Main layout with border container for better structure
	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		tabs,
		widget.NewSeparator(),
		statusSection,
	)
	
	// Add padding around the entire content
	return container.NewPadded(content)
}

// createFolderSection creates the folder selection section
func (a *GUIApp) createFolderSection() fyne.CanvasObject {
	// Excel folder with icon
	excelFolderBtn := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		a.selectFolder("Select Excel Folder", a.excelFolderEntry)
	})
	excelFolderRow := container.NewBorder(nil, nil, nil, excelFolderBtn, a.excelFolderEntry)
	
	// Schema folder with icon
	schemaFolderBtn := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		a.selectFolder("Select Schema Folder", a.schemaFolderEntry)
	})
	schemaFolderRow := container.NewBorder(nil, nil, nil, schemaFolderBtn, a.schemaFolderEntry)
	
	// Output folder with icon
	outputFolderBtn := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		a.selectFolder("Select Output Folder", a.outputFolderEntry)
	})
	outputFolderRow := container.NewBorder(nil, nil, nil, outputFolderBtn, a.outputFolderEntry)
	
	// Form with better spacing
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Excel Files", Widget: excelFolderRow, HintText: "Folder containing your Excel files"},
			{Text: "Schema Location", Widget: schemaFolderRow, HintText: "Where schema.yml will be saved"},
			{Text: "Output Location", Widget: outputFolderRow, HintText: "Where output.json will be saved"},
		},
	}
	
	// Card with icon
	card := widget.NewCard(
		"Configuration",
		"Select folders for processing your Excel files",
		form,
	)
	
	return card
}

// createActionsSection creates the actions section
func (a *GUIApp) createActionsSection() fyne.CanvasObject {
	// Action buttons with icons and importance styling
	generateBtn := widget.NewButtonWithIcon("Generate Schema", theme.DocumentCreateIcon(), a.generateSchema)
	generateBtn.Importance = widget.HighImportance
	
	updateBtn := widget.NewButtonWithIcon("Update Schema", theme.DocumentSaveIcon(), a.updateSchema)
	updateBtn.Importance = widget.MediumImportance
	
	dataBtn := widget.NewButtonWithIcon("Generate Data", theme.DownloadIcon(), a.generateData)
	dataBtn.Importance = widget.HighImportance
	
	// Button container with better spacing
	buttons := container.New(
		layout.NewGridLayoutWithColumns(3),
		generateBtn,
		updateBtn,
		dataBtn,
	)
	
	// Card with description
	card := widget.NewCard(
		"Actions",
		"Choose an operation to perform",
		buttons,
	)
	
	return card
}

// createStatusSection creates the status section
func (a *GUIApp) createStatusSection() *fyne.Container {
	// Status card for better visual grouping
	statusCard := widget.NewCard(
		"Status",
		"",
		container.NewVBox(
			a.statusLabel,
			a.progressBar,
		),
	)
	
	return container.NewVBox(statusCard)
}

// createHeader creates the application header
func (a *GUIApp) createHeader() fyne.CanvasObject {
	// App icon
	icon := widget.NewIcon(theme.ComputerIcon())
	
	// Title with larger text
	title := widget.NewLabelWithStyle(
		fmt.Sprintf("%s", a.name),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	
	// Version label
	version := widget.NewLabel(fmt.Sprintf("Version %s", a.version))
	version.Alignment = fyne.TextAlignCenter
	
	// Description
	description := widget.NewLabel("Convert Excel files to structured JSON data for Unity")
	description.Alignment = fyne.TextAlignCenter
	
	// Header container
	header := container.NewVBox(
		container.NewCenter(icon),
		title,
		version,
		description,
	)
	
	return container.NewCenter(header)
}

// selectFolder opens a folder selection dialog
func (a *GUIApp) selectFolder(title string, entry *widget.Entry) {
	dialog.ShowFolderOpen(func(reader fyne.ListableURI, err error) {
		if err != nil {
			a.showError(fmt.Sprintf("Error selecting folder: %v", err))
			return
		}
		if reader == nil {
			return // User cancelled
		}
		
		// Get the path from the URI
		path := reader.Path()
		
		// Enable entry temporarily to set text
		entry.Enable()
		entry.SetText(path)
		entry.Disable()
		
		a.logger.Debug("Folder selected", "title", title, "path", path)
		a.setStatus(fmt.Sprintf("Selected: %s", path))
	}, a.window)
}

// generateSchema handles schema generation
func (a *GUIApp) generateSchema() {
	a.logger.Info("Generate schema requested")
	
	folderPath := a.excelFolderEntry.Text
	if folderPath == "" {
		a.showError("Please select an Excel folder first")
		return
	}
	
	schemaPath := a.schemaFolderEntry.Text
	if schemaPath == "" {
		schemaPath = folderPath // Default to Excel folder
	}
	
	a.setStatus("Generating schema...")
	a.showProgress()
	
	// Disable buttons during operation
	a.disableActions()
	
	// Run in goroutine to avoid blocking UI
	go func() {
		defer func() {
			a.hideProgress()
			a.enableActions()
		}()
		
		ctx := context.Background()
		
		// Simulate progress updates
		a.progressBar.SetValue(0.2)
		a.setStatus("Scanning Excel files...")
		
		schema, err := a.schemaGenerator.GenerateFromFolder(ctx, folderPath)
		if err != nil {
			a.showError(fmt.Sprintf("Failed to generate schema: %v", err))
			return
		}
		
		a.progressBar.SetValue(0.8)
		a.setStatus("Saving schema...")
		
		// Save schema logic would go here
		
		a.progressBar.SetValue(1.0)
		a.setStatus(fmt.Sprintf("✓ Schema generated successfully with %d files", len(schema.Files)))
		a.showSuccess(fmt.Sprintf("Schema generated successfully!\n\nFound %d Excel files with %d total sheets.", 
			len(schema.Files), a.countSheets(schema)))
	}()
}

// updateSchema handles schema updates
func (a *GUIApp) updateSchema() {
	a.logger.Info("Update schema requested")
	
	if a.excelFolderEntry.Text == "" {
		a.showError("Please select an Excel folder first")
		return
	}
	
	a.showInfo("Update Schema", "This feature will update an existing schema.yml with any changes in your Excel files.\n\nComing soon!")
}

// generateData handles data generation
func (a *GUIApp) generateData() {
	a.logger.Info("Generate data requested")
	
	if a.excelFolderEntry.Text == "" {
		a.showError("Please select an Excel folder first")
		return
	}
	
	if a.schemaFolderEntry.Text == "" {
		a.showError("Please select a schema folder first")
		return
	}
	
	a.showInfo("Generate Data", "This feature will generate JSON data from your Excel files based on the schema.\n\nComing soon!")
}

// setStatus sets the status message
func (a *GUIApp) setStatus(message string) {
	a.statusLabel.SetText(message)
	a.logger.Debug("Status updated", "message", message)
}

// showProgress shows the progress bar
func (a *GUIApp) showProgress() {
	a.progressBar.Show()
}

// hideProgress hides the progress bar
func (a *GUIApp) hideProgress() {
	a.progressBar.Hide()
}

// showError shows an error message
func (a *GUIApp) showError(message string) {
	a.setStatus(fmt.Sprintf("Error: %s", message))
	a.logger.Error("GUI error", "message", message)
	
	dialog.ShowError(fmt.Errorf(message), a.window)
}

// showSuccess shows a success message
func (a *GUIApp) showSuccess(message string) {
	a.logger.Info("GUI success", "message", message)
	
	dialog.ShowInformation("Success", message, a.window)
}

// showInfo shows an information message
func (a *GUIApp) showInfo(title, message string) {
	a.logger.Info("GUI info", "title", title, "message", message)
	
	dialog.ShowInformation(title, message, a.window)
}

// disableActions disables action buttons
func (a *GUIApp) disableActions() {
	// This would disable the action buttons
	// Implementation depends on storing button references
}

// enableActions enables action buttons
func (a *GUIApp) enableActions() {
	// This would enable the action buttons
	// Implementation depends on storing button references
}

// countSheets counts total sheets in schema
func (a *GUIApp) countSheets(schemaInfo *models.SchemaInfo) int {
	count := 0
	if schemaInfo != nil && schemaInfo.Files != nil {
		for _, file := range schemaInfo.Files {
			count += len(file.Sheets)
		}
	}
	return count
}

// createSchemaTab creates the schema generation tab content
func (a *GUIApp) createSchemaTab() fyne.CanvasObject {
	// Folder selection section
	folderSection := a.createFolderSection()
	
	// Actions section
	actionsSection := a.createActionsSection()
	
	// Add some padding and spacing
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(0, 20))
	
	return container.NewVBox(
		folderSection,
		spacer,
		actionsSection,
	)
}

// createDriveDownloadTab creates the Google Drive download tab content
func (a *GUIApp) createDriveDownloadTab() fyne.CanvasObject {
	// Credentials file selection
	credentialsBtn := widget.NewButtonWithIcon("Browse", theme.FileIcon(), func() {
		a.selectFile("Select Credentials File", a.credentialsEntry, []string{".json"})
	})
	credentialsRow := container.NewBorder(nil, nil, nil, credentialsBtn, a.credentialsEntry)
	
	// Output folder selection
	outputBtn := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		a.selectFolder("Select Download Output Folder", a.downloadOutputEntry)
	})
	outputRow := container.NewBorder(nil, nil, nil, outputBtn, a.downloadOutputEntry)
	
	// Form
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Credentials File", Widget: credentialsRow, HintText: "Google Cloud credentials JSON file"},
			{Text: "Drive Folder Link", Widget: a.driveLinkEntry, HintText: "https://drive.google.com/drive/folders/..."},
			{Text: "Output Folder", Widget: outputRow, HintText: "Where to save downloaded files"},
		},
	}
	
	// Download button
	downloadBtn := widget.NewButtonWithIcon("Download from Drive", theme.DownloadIcon(), a.downloadFromDrive)
	downloadBtn.Importance = widget.HighImportance
	
	// Card
	card := widget.NewCard(
		"Google Drive Download",
		"Download Excel and Google Sheets files from a Google Drive folder",
		container.NewVBox(
			form,
			container.NewPadded(downloadBtn),
		),
	)
	
	return card
}

// selectFile opens a file selection dialog
func (a *GUIApp) selectFile(title string, entry *widget.Entry, filters []string) {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			a.showError(fmt.Sprintf("Error selecting file: %v", err))
			return
		}
		if reader == nil {
			return // User cancelled
		}
		defer reader.Close()
		
		// Get the path from the URI
		path := reader.URI().Path()
		
		// Enable entry temporarily to set text
		entry.Enable()
		entry.SetText(path)
		entry.Disable()
		
		a.logger.Debug("File selected", "title", title, "path", path)
		a.setStatus(fmt.Sprintf("Selected: %s", path))
	}, a.window)
}

// downloadFromDrive handles downloading files from Google Drive
func (a *GUIApp) downloadFromDrive() {
	a.logger.Info("Download from Drive requested")
	
	credentialsPath := a.credentialsEntry.Text
	if credentialsPath == "" {
		a.showError("Please select a credentials file first")
		return
	}
	
	driveLink := a.driveLinkEntry.Text
	if driveLink == "" {
		a.showError("Please enter a Google Drive folder link")
		return
	}
	
	outputPath := a.downloadOutputEntry.Text
	if outputPath == "" {
		a.showError("Please select an output folder")
		return
	}
	
	a.setStatus("Downloading from Google Drive...")
	a.showProgress()
	
	// Disable UI during operation
	a.disableActions()
	
	// Run in goroutine to avoid blocking UI
	go func() {
		defer func() {
			a.hideProgress()
			a.enableActions()
		}()
		
		ctx := context.Background()
		
		a.progressBar.SetValue(0.1)
		a.setStatus("Creating Google Drive client...")
		
		// Create downloader
		downloader, err := gdrive.NewDownloader(ctx, credentialsPath)
		if err != nil {
			a.showError(fmt.Sprintf("Failed to create downloader: %v", err))
			return
		}
		
		a.progressBar.SetValue(0.3)
		a.setStatus("Downloading files from Google Drive...")
		
		// Download files
		if err := downloader.DownloadFromDriveLink(driveLink, outputPath); err != nil {
			a.showError(fmt.Sprintf("Failed to download from Drive: %v", err))
			return
		}
		
		a.progressBar.SetValue(1.0)
		a.setStatus("✓ Download completed successfully")
		a.showSuccess(fmt.Sprintf("Successfully downloaded files to:\n%s", outputPath))
	}()
}