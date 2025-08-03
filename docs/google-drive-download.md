# Google Drive Download Feature

This feature allows you to download Excel files and Google Sheets from a Google Drive folder.

## Prerequisites

1. Google Cloud Service Account credentials JSON file
   - Create a project in Google Cloud Console
   - Enable Google Drive API
   - Create a Service Account credential
   - Download the Service Account JSON key file
   - Share the Google Drive folder with the Service Account email

## CLI Usage

```bash
./excel-schema-generator download \
  -credentials ./credentials.json \
  -drive-link 'https://drive.google.com/drive/folders/YOUR_FOLDER_ID' \
  -output ./downloads
```

### Parameters
- `-credentials`: Path to your Google credentials JSON file
- `-drive-link`: Google Drive folder URL
- `-output`: Local folder where files will be downloaded

## GUI Usage

1. Launch the application without arguments
2. Go to the "Google Drive Download" tab
3. Select your credentials JSON file
4. Paste the Google Drive folder link
5. Select output folder
6. Click "Download from Drive"

## Supported File Types

- Excel files (.xlsx, .xls)
- Google Sheets (automatically converted to .xlsx)

## Notes

- The tool will recursively download all Excel and Google Sheets files from the specified folder and its subfolders
- Google Sheets are automatically converted to Excel format during download
- Make sure to share the Google Drive folder with your Service Account email (found in the credentials JSON file)