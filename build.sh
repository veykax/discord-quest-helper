#!/usr/bin/env bash
set -e

echo "Building Discord Quest Helper..."
mkdir -p build
cp games.json internal/config/default_games.json

GO_CMD="${GO:-go}"
if ! command -v "$GO_CMD" >/dev/null 2>&1; then
    echo "Error: Go compiler ($GO_CMD) not found in PATH." >&2
    exit 1
fi

# Windows build (GUI subsystem, no console window, stripped)
echo "-> Building Windows binary (build/discord-quest-helper.exe)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 "$GO_CMD" build -ldflags="-s -w -H windowsgui" -o build/discord-quest-helper.exe

# Linux build (stripped)
echo "-> Building Linux binary (build/discord-quest-helper)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 "$GO_CMD" build -ldflags="-s -w" -o build/discord-quest-helper
chmod +x build/discord-quest-helper

echo "Done! Binaries generated in build/:"
ls -lh build/
