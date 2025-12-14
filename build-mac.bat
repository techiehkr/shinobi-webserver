# macOS Intel
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o build/macos/shinobi-webserver ./cmd/site-manager

# macOS Apple Silicon (M1/M2/M3)
set GOARCH=arm64
go build -ldflags="-s -w" -o build/macos/shinobi-webserver-arm64 ./cmd/site-manager