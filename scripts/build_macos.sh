#!/bin/bash
set -e

# Create bin directory if it doesn't exist
mkdir -p bin

# Build for Intel Macs
echo "Building for Intel Mac (amd64)..."
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o bin/excel-schema-generator-amd64 .

# Build for M1 Macs
echo "Building for M1 Mac (arm64)..."
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o bin/excel-schema-generator-arm64 .

# Create universal binary
echo "Creating universal binary..."
lipo -create -output bin/excel-schema-generator bin/excel-schema-generator-amd64 bin/excel-schema-generator-arm64

# Clean up intermediate files
rm bin/excel-schema-generator-amd64 bin/excel-schema-generator-arm64

echo "Build complete: bin/excel-schema-generator (universal binary)"