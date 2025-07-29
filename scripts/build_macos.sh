#!/bin/bash
set -e

# 创建 bin 目录（如果不存在）
mkdir -p bin

# 为 Intel Macs 构建
echo "Building for Intel Mac (amd64)..."
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o bin/excel-schema-generator-amd64 .

# 为 M1 Macs 构建
echo "Building for M1 Mac (arm64)..."
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o bin/excel-schema-generator-arm64 .

# 创建通用二进制文件
echo "Creating universal binary..."
lipo -create -output bin/excel-schema-generator bin/excel-schema-generator-amd64 bin/excel-schema-generator-arm64

# 清理中间文件
rm bin/excel-schema-generator-amd64 bin/excel-schema-generator-arm64

echo "Build complete: bin/excel-schema-generator (universal binary)"