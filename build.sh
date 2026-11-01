#!/usr/bin/env bash
set -e

echo "Building Discord Quest Helper..."
mkdir -p build

# Windows build (GUI subsystem, no console window, stripped)
echo "-> Building Windows binary (build/discord-quest-helper.exe)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 ~/.local/go/bin/go build -ldflags="-s -w -H windowsgui" -o build/discord-quest-helper.exe

# Linux build (stripped)
echo "-> Building Linux binary (build/discord-quest-helper)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 ~/.local/go/bin/go build -ldflags="-s -w" -o build/discord-quest-helper

echo "Done! Binaries generated in build/:"
ls -lh build/
