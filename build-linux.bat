# Linux 64-bit
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o build/linux/shinobi-webserver ./cmd/site-manager

# Linux ARM (Raspberry Pi)
set GOARCH=arm64
go build -ldflags="-s -w" -o build/linux/shinobi-webserver-arm64 ./cmd/site-manager

