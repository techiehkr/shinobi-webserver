# Create build directory
mkdir -p build\windows

# Build with different architectures
# 64-bit Windows (most common)
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o build/windows/shinobi-webserver.exe ./cmd/site-manager

# 32-bit Windows (older systems)
set GOARCH=386
go build -ldflags="-s -w" -o build/windows/shinobi-webserver-32bit.exe ./cmd/site-manager

