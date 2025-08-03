package gdrive

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type Downloader struct {
	service *drive.Service
	ctx     context.Context
}

func NewDownloader(ctx context.Context, credentialsFile string) (*Downloader, error) {
	client, err := getClient(ctx, credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	service, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create drive service: %w", err)
	}

	return &Downloader{
		service: service,
		ctx:     ctx,
	}, nil
}

func (d *Downloader) DownloadFromDriveLink(driveLink, outputFolder string) error {
	folderID, err := extractFolderID(driveLink)
	if err != nil {
		return fmt.Errorf("failed to extract folder ID: %w", err)
	}

	if err := os.MkdirAll(outputFolder, 0755); err != nil {
		return fmt.Errorf("failed to create output folder: %w", err)
	}

	return d.downloadFolder(folderID, outputFolder)
}

func (d *Downloader) downloadFolder(folderID, outputPath string) error {
	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	fileList, err := d.service.Files.List().Q(query).Fields("files(id, name, mimeType)").Do()
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	for _, file := range fileList.Files {
		switch file.MimeType {
		case "application/vnd.google-apps.folder":
			subFolderPath := filepath.Join(outputPath, file.Name)
			if err := os.MkdirAll(subFolderPath, 0755); err != nil {
				return fmt.Errorf("failed to create subfolder %s: %w", file.Name, err)
			}
			if err := d.downloadFolder(file.Id, subFolderPath); err != nil {
				return fmt.Errorf("failed to download subfolder %s: %w", file.Name, err)
			}

		case "application/vnd.google-apps.spreadsheet":
			outputFile := filepath.Join(outputPath, file.Name+".xlsx")
			if err := d.downloadGoogleSheet(file.Id, outputFile); err != nil {
				return fmt.Errorf("failed to download Google Sheet %s: %w", file.Name, err)
			}
			fmt.Printf("Downloaded Google Sheet: %s\n", outputFile)

		case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"application/vnd.ms-excel":
			outputFile := filepath.Join(outputPath, file.Name)
			if err := d.downloadFile(file.Id, outputFile); err != nil {
				return fmt.Errorf("failed to download Excel file %s: %w", file.Name, err)
			}
			fmt.Printf("Downloaded Excel file: %s\n", outputFile)
		}
	}

	return nil
}

func (d *Downloader) downloadGoogleSheet(fileID, outputPath string) error {
	resp, err := d.service.Files.Export(fileID, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet").Download()
	if err != nil {
		return fmt.Errorf("failed to export Google Sheet: %w", err)
	}
	defer resp.Body.Close()

	return saveResponseToFile(resp, outputPath)
}

func (d *Downloader) downloadFile(fileID, outputPath string) error {
	resp, err := d.service.Files.Get(fileID).Download()
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	return saveResponseToFile(resp, outputPath)
}

func saveResponseToFile(resp *http.Response, outputPath string) error {
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

func extractFolderID(driveLink string) (string, error) {
	// Handle various Google Drive URL formats
	// Example: https://drive.google.com/drive/folders/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms
	if idx := strings.Index(driveLink, "/folders/"); idx != -1 {
		parts := strings.Split(driveLink[idx+9:], "?")
		if len(parts[0]) > 0 {
			return parts[0], nil
		}
	}
	
	// Handle format with id parameter
	if idx := strings.Index(driveLink, "id="); idx != -1 {
		parts := strings.Split(driveLink[idx+3:], "&")
		if len(parts[0]) > 0 {
			return parts[0], nil
		}
	}
	
	// Handle file format: https://drive.google.com/file/d/FILE_ID/view
	if idx := strings.Index(driveLink, "/d/"); idx != -1 {
		endIdx := strings.Index(driveLink[idx+3:], "/")
		if endIdx == -1 {
			return driveLink[idx+3:], nil
		}
		return driveLink[idx+3 : idx+3+endIdx], nil
	}

	return "", fmt.Errorf("could not extract folder ID from link: %s", driveLink)
}