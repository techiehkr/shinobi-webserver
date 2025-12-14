#!/bin/bash

echo "Building Shinobi Web Server for all platforms..."
echo "==============================================="

# Windows
echo ""
echo "[1/3] Building for Windows..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o build/windows/shinobi-webserver.exe ./cmd/site-manager

# Linux
echo ""
echo "[2/3] Building for Linux..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o build/linux/shinobi-webserver ./cmd/site-manager

# macOS
echo ""
echo "[3/3] Building for macOS..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o build/macos/shinobi-webserver ./cmd/site-manager

echo ""
echo "✅ Build complete!"
echo ""
echo "Windows: build/windows/shinobi-webserver.exe"
echo "Linux:   build/linux/shinobi-webserver"
echo "macOS:   build/macos/shinobi-webserver"
echo ""
ls -lh build/windows/shinobi-webserver.exe
ls -lh build/linux/shinobi-webserver
ls -lh build/macos/shinobi-webserver