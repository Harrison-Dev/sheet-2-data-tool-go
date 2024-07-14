#!/bin/bash
set -e

# 为 Intel Macs 构建
echo "Building for Intel Mac (amd64)..."
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o data-generator-amd64 .

# 为 M1 Macs 构建
echo "Building for M1 Mac (arm64)..."
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o data-generator-arm64 .

# 创建通用二进制文件
echo "Creating universal binary..."
lipo -create -output data-generator data-generator-amd64 data-generator-arm64

# 清理中间文件
rm data-generator-amd64 data-generator-arm64

echo "Build complete: data-generator (universal binary)"