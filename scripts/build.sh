#!/bin/bash
set -e

# Get the directory of the script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Read version from VERSION file
VERSION=$(cat "$ROOT_DIR/VERSION")

echo "Building sortpath v$VERSION..."

# Create bin directory
mkdir -p "$ROOT_DIR/bin"

# Build for current platform
go build \
  -ldflags="-X main.Version=$VERSION" \
  -o "$ROOT_DIR/bin/sortpath" \
  "$ROOT_DIR/cmd/sortpath.go"

echo "Build complete: $ROOT_DIR/bin/sortpath"
echo "Version: $VERSION"

# Make executable
chmod +x "$ROOT_DIR/bin/sortpath"