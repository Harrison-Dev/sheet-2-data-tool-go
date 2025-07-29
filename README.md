# Excel Schema Generator

A tool for extracting schema definitions from Excel files and generating JSON data. Supports both Command Line Interface (CLI) and Graphical User Interface (GUI) operation modes.

## Features

- **Automatic Schema Extraction**: Automatically analyzes Excel files and generates YAML format schema definitions
- **Data Conversion**: Converts Excel data to JSON format based on schema definitions
- **Dual Operation Modes**:
  - CLI mode: Suitable for automation and batch processing
  - GUI mode: Provides a user-friendly visual interface
- **Schema Updates**: Supports incremental updates to existing schema definitions
- **Cross-platform Support**: Supports Windows, macOS (Intel and Apple Silicon)

## Installation

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
```

## Usage

### GUI Mode

Launch the graphical interface by running the program without arguments:

```bash
./data-generator
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
./data-generator generate -folder /path/to/excel/files
```

This will generate a `schema.yml` file in the current directory.

#### 2. Update Schema

Update existing schema when Excel files have changed:

```bash
./data-generator update -folder /path/to/excel/files
```

#### 3. Generate JSON Data

Extract data from Excel files based on schema:

```bash
./data-generator data -folder /path/to/excel/files
```

This will generate an `output.json` file in the current directory.

## Workflow

1. **Initialize**: Use the `generate` command to create initial schema
2. **Customize**: Edit `schema.yml` to adjust data types and field names
3. **Update**: Use the `update` command when Excel structure changes
4. **Output**: Use the `data` command to generate final JSON data

## Schema Format

Example `schema.yml` file structure:

```yaml
files:
  example.xlsx:
    sheets:
      Sheet1:
        offset_header: 0
        class_name: "ExampleData"
        sheet_name: "Sheet1"
        data_class:
          - name: "id"
            data_type: "string"
          - name: "name"
            data_type: "string"
          - name: "value"
            data_type: "number"
```

### Field Descriptions

- `offset_header`: Offset of the header row (0 means first row)
- `class_name`: Data class name
- `sheet_name`: Excel sheet name
- `data_class`: Field definition list
  - `name`: Field name
  - `data_type`: Data type (string, number, boolean, etc.)

## Building the Project

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

## Project Structure

```
sheet-2-data-tool-go/
├── main.go              # Main entry point
├── gui.go               # GUI implementation
├── config.go            # Configuration management
├── excelschema/         # Core functionality package
│   ├── models.go        # Data structure definitions
│   ├── generate-schema.go # Schema generation logic
│   ├── update-schema.go   # Schema update logic
│   └── generate-data.go   # Data generation logic
└── scripts/             # Build scripts
    ├── build_macos.sh
    └── build_windows.bat
```

## Dependencies

- [fyne.io/fyne/v2](https://fyne.io/) - GUI framework
- [github.com/xuri/excelize/v2](https://github.com/qax-os/excelize) - Excel file processing
- [gopkg.in/yaml.v2](https://gopkg.in/yaml.v2) - YAML parsing

## License

[Please add your license information]

## Contributing

Issues and Pull Requests are welcome!

## Author

[Please add author information]