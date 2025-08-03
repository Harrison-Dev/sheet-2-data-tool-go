# Excel Schema Generator v2.0

A modern, refactored Excel Schema Generator built with clean architecture principles. This tool processes Excel files to generate YAML schemas and JSON data outputs, specifically designed for Unity game development workflows.

## ✨ Key Features

- **Clean Architecture**: Follows hexagonal/clean architecture patterns for better maintainability
- **Dual Interface**: Both CLI and GUI modes (GUI coming soon in v2.0)
- **Cross-Platform**: Supports Windows, macOS, and Linux
- **Comprehensive Logging**: Structured logging with configurable levels
- **Error Handling**: Robust error handling with user-friendly messages
- **Schema Management**: Generate, update, and validate Excel schemas
- **Data Generation**: Export Excel data as structured JSON
- **Unity Compatible**: Output format designed for Unity master memory project

## 🏗️ Architecture

The application follows clean architecture principles with the following layers:

### Core Components
- **Domain Layer** (`internal/core/models/`): Business entities and rules
- **Application Layer** (`internal/core/`): Use cases and business logic
- **Infrastructure Layer** (`internal/adapters/`): External system integrations
- **Ports** (`internal/ports/`): Interface definitions

### Key Services
- **Schema Service**: Manages schema generation, updates, and validation
- **Excel Service**: Handles Excel file processing and data extraction
- **Validation Service**: Validates schemas and data integrity
- **File Service**: Manages file system operations

## 🚀 Installation

### Prerequisites
- Go 1.21 or later
- Excel files (.xlsx or .xls format)

### Build from Source
```bash
git clone <repository-url>
cd excel-schema-generator
go build .
```

### Platform-Specific Builds
```bash
# macOS (universal binary)
./scripts/build/build_macos.sh

# Windows
./scripts/build/build_windows.bat

# Linux
./scripts/build/build_linux.sh
```

## 📖 Usage

### CLI Commands

#### Generate Schema
Create a new schema or update existing schema from Excel files:
```bash
./excel-schema-generator generate -folder ./excel-files [-output ./schemas]
```
*Note: Automatically detects if schema.yml exists and will create new or update accordingly while preserving manual field settings.*

#### Generate Data
Generate JSON data from Excel files using an existing schema:
```bash
./excel-schema-generator data -folder ./excel-files [-output ./output]
```

### Command Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-folder` | Path to Excel files folder (required) | - |
| `-output` | Output directory path | Current directory |
| `-verbose` | Enable verbose logging | false |
| `-log-level` | Log level (debug, info, warn, error) | info |
| `-log-format` | Log format (text, json) | text |

### GUI Mode
```bash
./excel-schema-generator
```
*Note: GUI mode is coming soon in v2.0*

## 📁 File Formats

### Input Files
- **Excel Files**: `.xlsx` and `.xls` formats
- **Schema Files**: `schema.yml` (YAML format)

### Output Files
- **Schema**: `schema.yml` - Contains Excel structure definitions
- **Data**: `output.json` - Generated JSON data for Unity

### Schema Structure
```yaml
version: "1.0"
metadata:
  description: "Generated Excel schema for data conversion"
  schema_version: "1.0"
created_at: 2024-01-01T00:00:00Z
updated_at: 2024-01-01T00:00:00Z
files:
  "example.xlsx":
    file_name: "example.xlsx"
    file_path: "example.xlsx"
    checksum: "abc123..."
    last_updated: 2024-01-01T00:00:00Z
    sheets:
      "Sheet1":
        sheet_name: "Sheet1"
        class_name: "Sheet1"
        offset_header: 1
        row_count: 100
        data_class:
          - name: "Id"
            data_type: "int"
            required: true
          - name: "Name"
            data_type: "string"
            required: false
```

### Output Data Structure
```json
{
  "metadata": {
    "generated_at": "2024-01-01T00:00:00Z",
    "schema_version": "1.0",
    "generator": "Excel Schema Generator v2.0",
    "file_count": 1,
    "record_count": 100
  },
  "schema": {
    "Sheet1": [
      {
        "name": "Id",
        "dataType": "int"
      },
      {
        "name": "Name",
        "dataType": "string"
      }
    ]
  },
  "data": {
    "Sheet1": [
      {
        "Id": 1,
        "Name": "Example Item"
      }
    ]
  }
}
```

## 🔧 Configuration

### Logging Configuration
Control logging behavior through command-line flags:

```bash
# Debug level with JSON format
./excel-schema-generator generate -folder ./data -log-level debug -log-format json

# Verbose mode
./excel-schema-generator generate -folder ./data -verbose
```

### Excel Processing Options
The application automatically handles:
- Multiple sheets per Excel file
- Header row detection (configurable offset)
- Data type inference
- Empty row/column handling
- Temporary file exclusion (`~$` prefix files)

## 🧪 Development

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test packages
go test ./internal/core/schema/...
```

### Code Quality
```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Vet code
go vet ./...
```

### Project Structure
```
excel-schema-generator/
├── cmd/                    # Application entry points
│   ├── cli/               # CLI application
│   └── gui/               # GUI application (coming soon)
├── internal/              # Private application packages
│   ├── core/              # Business logic
│   │   ├── models/        # Domain models
│   │   └── schema/        # Schema services
│   ├── adapters/          # Infrastructure adapters
│   │   ├── excel/         # Excel processing
│   │   └── filesystem/    # File system operations
│   ├── ports/             # Interface definitions
│   └── utils/             # Shared utilities
│       ├── errors/        # Error handling
│       ├── logger/        # Logging utilities
│       └── validation/    # Validation services
├── pkg/                   # Public packages
│   └── logger/            # Logger implementation
├── test/                  # Test files and fixtures
└── scripts/               # Build and utility scripts
```

## 🚨 Error Handling

The application provides comprehensive error handling with user-friendly messages:

### Common Errors
- **File Not Found**: Occurs when Excel files or schema files don't exist
- **Invalid Format**: Occurs when Excel files are corrupted or unsupported
- **Validation Errors**: Occurs when schema or data validation fails
- **Permission Errors**: Occurs when file access is denied

### Error Categories
- `VALIDATION_ERROR`: Input validation failures
- `FILE_ERROR`: File system operation failures
- `EXCEL_ERROR`: Excel processing failures
- `SCHEMA_ERROR`: Schema-related failures
- `CONFIG_ERROR`: Configuration issues

## 🔄 Migration from v1.x

The v2.0 refactoring maintains backward compatibility for:
- CLI command interface (`generate`, `update`, `data`)
- Output file formats (`schema.yml`, `output.json`)
- Unity integration patterns

### Breaking Changes
- GUI mode is temporarily unavailable (coming soon)
- The `update` command has been integrated into `generate` command (v0.1.0+)
- Some internal APIs have changed (not affecting CLI usage)
- Log format has been standardized

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes following the architecture patterns
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit changes (`git commit -m 'Add amazing feature'`)
7. Push to branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Guidelines
- Follow clean architecture principles
- Add comprehensive tests for new features
- Use structured logging throughout
- Document public APIs
- Handle errors gracefully

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Excelize](https://github.com/xuri/excelize) for Excel processing
- Uses [Fyne](https://fyne.io/) for GUI framework (coming soon)
- Structured logging with Go's `slog` package
- Clean architecture principles inspired by Robert C. Martin

## 📞 Support

- Create an issue for bug reports or feature requests
- Check existing issues before creating new ones
- Provide detailed information including OS, Go version, and error messages

---

**Excel Schema Generator v2.0** - Transforming Excel data for modern applications with clean, maintainable code.