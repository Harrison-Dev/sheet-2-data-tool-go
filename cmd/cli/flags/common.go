package flags

import (
	"flag"
)

// CommonFlags defines common flags used across CLI commands
type CommonFlags struct {
	FolderPath string
	OutputPath string
	Verbose    bool
	LogLevel   string
	LogFormat  string
}

// AddCommonFlags adds common flags to a flag set
func AddCommonFlags(fs *flag.FlagSet, flags *CommonFlags) {
	fs.StringVar(&flags.FolderPath, "folder", "", "Path to the Excel files folder")
	fs.StringVar(&flags.OutputPath, "output", "", "Path to the output directory (optional, defaults to current working directory)")
	fs.BoolVar(&flags.Verbose, "verbose", false, "Enable verbose logging")
	fs.StringVar(&flags.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	fs.StringVar(&flags.LogFormat, "log-format", "text", "Log format (text, json)")
}

// Validate validates common flags
func (f *CommonFlags) Validate() error {
	if f.FolderPath == "" {
		return &ValidationError{Field: "folder", Message: "folder path is required"}
	}
	return nil
}

// ValidationError represents a flag validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}