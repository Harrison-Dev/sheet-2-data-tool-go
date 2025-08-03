package gdrive

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

func getClient(ctx context.Context, credentialsFile string) (*http.Client, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %w", err)
	}

	// Use service account credentials
	config, err := google.JWTConfigFromJSON(b, drive.DriveReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse service account credentials: %w", err)
	}

	return config.Client(ctx), nil
}