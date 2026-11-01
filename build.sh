#!/usr/bin/env bash
set -e

echo "Building Discord Quest Helper..."

# Windows build (GUI subsystem, no console window, stripped)
echo "-> Building Windows binary (discord-quest-helper.exe)..."
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H windowsgui" -o discord-quest-helper.exe

# Linux build (stripped)
echo "-> Building Linux binary (discord-quest-helper)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o discord-quest-helper

echo "Done! Binaries generated:"
ls -lh discord-quest-helper*
