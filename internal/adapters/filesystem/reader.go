package filesystem

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"excel-schema-generator/internal/ports"
	"excel-schema-generator/internal/utils/errors"
)

// FileRepository implements the FileRepository interface
type FileRepository struct {
	logger ports.LoggingService
}

// NewFileRepository creates a new file repository
func NewFileRepository(logger ports.LoggingService) *FileRepository {
	return &FileRepository{
		logger: logger,
	}
}

// List lists files in a directory with optional pattern matching
func (r *FileRepository) List(ctx context.Context, dir string, pattern string) ([]string, error) {
	r.logger.Debug("Listing files", "directory", dir, "pattern", pattern)

	// Check if directory exists
	if exists, err := r.Exists(ctx, dir); err != nil {
		return nil, err
	} else if !exists {
		return nil, errors.NewFileError(errors.DirectoryNotFoundCode, fmt.Sprintf("Directory not found: %s", dir))
	}

	// Check if it's actually a directory
	if isDir, err := r.IsDir(ctx, dir); err != nil {
		return nil, err
	} else if !isDir {
		return nil, errors.NewFileError(errors.DirectoryNotFoundCode, fmt.Sprintf("Path is not a directory: %s", dir))
	}

	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			r.logger.Warn("Error walking directory", "path", path, "error", err)
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Apply pattern matching if specified
		if pattern != "" {
			matched, err := filepath.Match(pattern, info.Name())
			if err != nil {
				r.logger.Warn("Invalid pattern", "pattern", pattern, "error", err)
				return nil
			}
			if !matched {
				return nil
			}
		}

		// Calculate relative path
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			r.logger.Warn("Failed to calculate relative path", "path", path, "dir", dir, "error", err)
			relPath = path
		}

		files = append(files, relPath)
		return nil
	})

	if err != nil {
		return nil, errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to list directory contents")
	}

	r.logger.Debug("Listed files", "count", len(files))
	return files, nil
}

// Exists checks if a file or directory exists
func (r *FileRepository) Exists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Cannot check file existence")
}

// IsDir checks if a path is a directory
func (r *FileRepository) IsDir(ctx context.Context, path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, errors.NewFileError(errors.FileNotFoundCode, fmt.Sprintf("Path not found: %s", path))
		}
		return false, errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Cannot access path")
	}
	return info.IsDir(), nil
}

// GetInfo retrieves file information
func (r *FileRepository) GetInfo(ctx context.Context, path string) (*ports.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewFileError(errors.FileNotFoundCode, fmt.Sprintf("File not found: %s", path))
		}
		return nil, errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Cannot access file")
	}

	return &ports.FileInfo{
		Name:         info.Name(),
		Size:         info.Size(),
		IsDirectory:  info.IsDir(),
		LastModified: info.ModTime().Unix(),
		Path:         path,
	}, nil
}

// Read reads a file and returns its content
func (r *FileRepository) Read(ctx context.Context, path string) ([]byte, error) {
	r.logger.Debug("Reading file", "path", path)

	// Check if file exists
	if exists, err := r.Exists(ctx, path); err != nil {
		return nil, err
	} else if !exists {
		return nil, errors.NewFileError(errors.FileNotFoundCode, fmt.Sprintf("File not found: %s", path))
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to read file")
	}

	r.logger.Debug("File read successfully", "path", path, "size", len(content))
	return content, nil
}

// Write writes content to a file
func (r *FileRepository) Write(ctx context.Context, path string, content []byte) error {
	r.logger.Debug("Writing file", "path", path, "size", len(content))

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := r.CreateDir(ctx, dir, 0755); err != nil {
		return err
	}

	// Write file
	err := os.WriteFile(path, content, 0644)
	if err != nil {
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to write file")
	}

	r.logger.Debug("File written successfully", "path", path)
	return nil
}

// Copy copies a file from source to destination
func (r *FileRepository) Copy(ctx context.Context, src, dst string) error {
	r.logger.Debug("Copying file", "src", src, "dst", dst)

	// Check if source exists
	if exists, err := r.Exists(ctx, src); err != nil {
		return err
	} else if !exists {
		return errors.NewFileError(errors.FileNotFoundCode, fmt.Sprintf("Source file not found: %s", src))
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Cannot open source file")
	}
	defer srcFile.Close()

	// Create destination directory if needed
	dstDir := filepath.Dir(dst)
	if err := r.CreateDir(ctx, dstDir, 0755); err != nil {
		return err
	}

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Cannot create destination file")
	}
	defer dstFile.Close()

	// Copy content
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to copy file content")
	}

	r.logger.Debug("File copied successfully", "src", src, "dst", dst)
	return nil
}

// Delete removes a file or directory
func (r *FileRepository) Delete(ctx context.Context, path string) error {
	r.logger.Debug("Deleting path", "path", path)

	// Check if path exists
	if exists, err := r.Exists(ctx, path); err != nil {
		return err
	} else if !exists {
		// Already deleted, consider it successful
		return nil
	}

	// Remove file or directory
	err := os.RemoveAll(path)
	if err != nil {
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to delete path")
	}

	r.logger.Debug("Path deleted successfully", "path", path)
	return nil
}

// CreateDir creates a directory with the given permissions
func (r *FileRepository) CreateDir(ctx context.Context, path string, perm uint32) error {
	// Check if directory already exists
	if exists, err := r.Exists(ctx, path); err != nil {
		return err
	} else if exists {
		// Directory already exists, check if it's actually a directory
		if isDir, err := r.IsDir(ctx, path); err != nil {
			return err
		} else if !isDir {
			return errors.NewFileError(errors.FilePermissionCode, fmt.Sprintf("Path exists but is not a directory: %s", path))
		}
		return nil // Directory already exists
	}

	// Create directory
	err := os.MkdirAll(path, os.FileMode(perm))
	if err != nil {
		return errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to create directory")
	}

	r.logger.Debug("Directory created successfully", "path", path)
	return nil
}

// GetExcelFiles returns a list of Excel files in a directory
func (r *FileRepository) GetExcelFiles(ctx context.Context, dir string) ([]string, error) {
	r.logger.Debug("Getting Excel files", "directory", dir)

	var excelFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			r.logger.Warn("Error walking directory", "path", path, "error", err)
			return nil // Continue walking
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's an Excel file
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if ext == ".xlsx" || ext == ".xls" {
			// Skip temporary files
			if strings.HasPrefix(info.Name(), "~$") {
				r.logger.Debug("Skipping temporary Excel file", "file", info.Name())
				return nil
			}

			// Calculate relative path
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				r.logger.Warn("Failed to calculate relative path", "path", path, "dir", dir, "error", err)
				relPath = path
			}

			excelFiles = append(excelFiles, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, errors.WrapError(err, errors.FileErrorType, errors.FilePermissionCode, "Failed to scan for Excel files")
	}

	r.logger.Debug("Found Excel files", "count", len(excelFiles))
	return excelFiles, nil
}