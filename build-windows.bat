# Build with static linking to include all dependencies
$env:CGO_ENABLED = "1"
$env:CC = "gcc"
$env:CGO_LDFLAGS = "-static"

go build -ldflags="-s -w" -o build/windows/shinobi-static.exe ./cmd/site-manager

# Test the static version
.\build\windows\shinobi-static.exe