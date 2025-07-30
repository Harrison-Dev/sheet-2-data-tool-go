# Excel Schema Generator

A tool for extracting schema definitions from Excel files and generating JSON data. Supports both Command Line Interface (CLI) and Graphical User Interface (GUI) operation modes, with structured logging capabilities.

## ✨ Features

- **Automatic Schema Extraction**: Automatically analyzes Excel files and generates YAML format schema definitions
- **Data Conversion**: Converts Excel data to JSON format based on schema definitions
- **Dual Operation Modes**:
  - CLI mode: Suitable for automation and scripting
  - GUI mode: Provides a user-friendly visual interface
- **Structured Logging**: Built-in structured logging with configurable levels and formats
- **Schema Updates**: Supports incremental updates to existing schema definitions
- **Cross-platform Support**: Supports Windows, macOS (Intel and Apple Silicon)

## 🚀 Installation

### Download Pre-compiled Binaries
Download the version suitable for your operating system from the [Releases](https://github.com/yourusername/sheet-2-data-tool-go/releases) page.

### Build from Source

Requirements:
- Go 1.19 or higher
- CGO support (required for GUI mode)

```bash
# Clone the repository
git clone https://github.com/yourusername/sheet-2-data-tool-go.git
cd sheet-2-data-tool-go

# Build
go build .

# Or use the build scripts
./scripts/build_macos.sh      # For macOS (creates universal binary)
scripts\build_windows.bat     # For Windows
```

## 📖 Usage

### GUI Mode

Launch the graphical interface by running the program without arguments:

```bash
./excel-schema-generator
```

In the GUI, you can:
1. Select the folder containing Excel files
2. Specify the location to save schema definition files
3. Set the JSON output folder
4. Click buttons to execute corresponding operations

### CLI Mode

#### 1. Generate Initial Schema

Generate basic schema definition from Excel folder:

```bash
./excel-schema-generator generate -folder /path/to/excel/files [OPTIONS]
```

This will scan all Excel files in the specified folder and generate a `schema.yml` file in the current directory.

#### 2. Update Schema

Update existing schema when Excel files have changed:

```bash
./excel-schema-generator update -folder /path/to/excel/files [OPTIONS]
```

This will update the existing `schema.yml` file with any new columns or sheets found in the Excel files.

#### 3. Generate JSON Data

Extract data from Excel files based on schema:

```bash
./excel-schema-generator data -folder /path/to/excel/files [OPTIONS]
```

This will generate an `output.json` file containing all the data from the Excel files according to the schema definition.

**Common Options:**
- `-verbose`: Enable verbose logging
- `-log-level`: Set log level (debug, info, warn, error) (default: "info")
- `-log-format`: Set log format (text, json) (default: "text")

**Examples:**

```bash
# Generate schema with debug logging
./excel-schema-generator generate -folder ./excel_files -log-level debug -verbose

# Update schema with JSON format logging
./excel-schema-generator update -folder ./excel_files -log-format json

# Generate data with error-level logging only
./excel-schema-generator data -folder ./excel_files -log-level error
```

## 🔄 Workflow

1. **Initialize**: Use the `generate` command to create initial schema from your Excel files
2. **Customize**: Edit `schema.yml` to adjust data types and field names
3. **Update**: Use the `update` command when Excel structure changes
4. **Output**: Use the `data` command to generate final JSON data

## 📋 Schema Format

Example `schema.yml` file structure:

```yaml
files:
  example.xlsx:
    sheets:
      Sheet1:
        offset_header: 1        # Header row position (1-based)
        class_name: "ExampleData"
        sheet_name: "Sheet1"
        data_class:
          - name: "Id"          # Must have an "Id" field of type "int"
            data_type: "int"
          - name: "name"
            data_type: "string"
          - name: "value"
            data_type: "float"
          - name: "active"
            data_type: "bool"
```

### Field Descriptions

- `offset_header`: Position of the header row (1-based indexing)
- `class_name`: Data class name for the sheet
- `sheet_name`: Excel sheet name
- `data_class`: Field definition list
  - `name`: Field name (case-sensitive)
  - `data_type`: Data type (string, int, float, bool)

**Important**: Each sheet must have an "Id" field with data_type "int" for data generation to work properly.

## 🔧 Advanced Features

### Structured Logging

The application features comprehensive structured logging:

- **Log Levels**: Debug, Info, Warn, Error
- **Log Formats**: Text (human-readable) or JSON (machine-parseable)
- **Contextual Information**: All log entries include relevant context like file names, sheet names, etc.

Example log output (text format):
```
time=2025-07-30T10:15:30.123+08:00 level=INFO msg="Schema generation completed" file=schema.yml
```

Example log output (JSON format):
```json
{"time":"2025-07-30T10:15:30.123+08:00","level":"INFO","msg":"Schema generation completed","file":"schema.yml"}
```

## 🏗️ Building the Project

### macOS

```bash
# Build universal binary (supports Intel and Apple Silicon)
./scripts/build_macos.sh
```

### Windows

```bash
# Build Windows executable
scripts\build_windows.bat
```

### Development Build

```bash
go build .
```

## 🧪 Testing

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test ./... -v
```

Run specific package tests:

```bash
go test ./excelschema -v
go test ./pkg/logger -v
```

## 📁 Project Structure

```
sheet-2-data-tool-go/
├── main.go                    # Main entry point with CLI support
├── gui.go                     # GUI implementation
├── config.go                  # Configuration management
├── excelschema/               # Core functionality package
│   ├── models.go              # Data structure definitions
│   ├── generate-schema.go     # Schema generation logic
│   ├── update-schema.go       # Schema update logic
│   ├── generate-data.go       # Data generation logic
│   └── *_test.go              # Comprehensive test suite
├── pkg/                       # Additional packages
│   └── logger/                # Structured logging system
│       ├── logger.go
│       └── logger_test.go
└── scripts/                   # Build scripts
    ├── build_macos.sh
    └── build_windows.bat
```

## 📦 Dependencies

- [fyne.io/fyne/v2](https://fyne.io/) - GUI framework
- [github.com/xuri/excelize/v2](https://github.com/qax-os/excelize) - Excel file processing
- [gopkg.in/yaml.v2](https://gopkg.in/yaml.v2) - YAML parsing
- Built-in `log/slog` - Structured logging (Go 1.19+)

## 🔍 Troubleshooting

### Common Issues

1. **"No ID field found" Error**: Ensure each sheet has an "Id" column with data_type "int" in the schema
2. **Permission Errors**: Ensure the application has read/write permissions for the specified directories
3. **Empty Output**: Check that the offset_header value correctly points to your header row

### Tips

1. **Large Files**: For better performance with large Excel files, use appropriate log levels (avoid debug in production)
2. **Schema Validation**: Always review the generated schema.yml before generating data
3. **Data Types**: Ensure data types in schema match the actual data in Excel files

## 📄 License

[Please add your license information]

## 🤝 Contributing

Issues and Pull Requests are welcome!

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 👨‍💻 Author

[Please add author information]

## 📝 Changelog

### v1.1.0 (Latest)
- ✨ Added structured logging with configurable levels and formats
- 🧪 Comprehensive test coverage for all features
- 📚 Improved documentation
- 🔧 Better error handling and logging

### v1.0.0
- 🎉 Initial release
- ✨ Excel to YAML schema generation
- ✨ Schema update functionality
- ✨ JSON data generation from Excel
- ✨ GUI and CLI interfaces