#!/usr/bin/env bash

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="kacperkwapisz/sortpath"
BINARY_NAME="sortpath"
INSTALL_DIR="/usr/local/bin"
FALLBACK_DIR="$HOME/.local/bin"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os arch

    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        *)              
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac

    echo "${os}-${arch}"
}

# Get latest release version from GitHub API
get_latest_version() {
    local version
    log_info "Fetching latest release version..."
    
    if command -v curl >/dev/null 2>&1; then
        version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
    elif command -v wget >/dev/null 2>&1; then
        version=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    if [ -z "$version" ]; then
        log_error "Failed to fetch latest version"
        exit 1
    fi
    
    echo "$version"
}

# Download binary
download_binary() {
    local version="$1"
    local platform="$2"
    local binary_suffix=""
    local download_url
    local temp_file
    
    # Add .exe suffix for Windows
    if [[ "$platform" == "windows-"* ]]; then
        binary_suffix=".exe"
    fi
    
    download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-${platform}${binary_suffix}"
    temp_file="/tmp/${BINARY_NAME}-${platform}${binary_suffix}"
    
    log_info "Downloading ${BINARY_NAME} ${version} for ${platform}..."
    log_info "URL: ${download_url}"
    
    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL "$download_url" -o "$temp_file"; then
            log_error "Failed to download binary"
            exit 1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -q "$download_url" -O "$temp_file"; then
            log_error "Failed to download binary"
            exit 1
        fi
    else
        log_error "Neither curl nor wget is available"
        exit 1
    fi
    
    echo "$temp_file"
}

# Install binary
install_binary() {
    local temp_file="$1"
    local install_path
    
    # Try primary install directory first
    if [ -w "$INSTALL_DIR" ] || [ -w "$(dirname "$INSTALL_DIR")" ]; then
        install_path="$INSTALL_DIR/$BINARY_NAME"
    else
        # Try with sudo
        if sudo -n true 2>/dev/null; then
            log_info "Installing to $INSTALL_DIR (requires sudo)..."
            if sudo cp "$temp_file" "$INSTALL_DIR/$BINARY_NAME" && sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"; then
                install_path="$INSTALL_DIR/$BINARY_NAME"
            else
                log_warn "Failed to install to $INSTALL_DIR with sudo, trying fallback..."
                install_path=""
            fi
        else
            log_warn "No write permission to $INSTALL_DIR and sudo not available, using fallback..."
            install_path=""
        fi
    fi
    
    # Use fallback directory if primary failed
    if [ -z "$install_path" ]; then
        mkdir -p "$FALLBACK_DIR"
        install_path="$FALLBACK_DIR/$BINARY_NAME"
        log_info "Installing to $FALLBACK_DIR..."
        cp "$temp_file" "$install_path"
    fi
    
    # Make executable
    chmod +x "$install_path"
    
    echo "$install_path"
}

# Check if binary is in PATH
check_path() {
    local install_path="$1"
    local install_dir
    install_dir="$(dirname "$install_path")"
    
    if [[ ":$PATH:" != *":$install_dir:"* ]]; then
        log_warn "$install_dir is not in your PATH"
        log_info "Add it to your PATH by adding this line to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
        echo "export PATH=\"$install_dir:\$PATH\""
        echo ""
        log_info "Or run the following command to add it temporarily:"
        echo "export PATH=\"$install_dir:\$PATH\""
        echo ""
    fi
}

# Cleanup
cleanup() {
    local temp_file="$1"
    if [ -f "$temp_file" ]; then
        rm -f "$temp_file"
    fi
}

# Main installation function
main() {
    local platform version temp_file install_path
    
    log_info "Starting ${BINARY_NAME} installation..."
    
    # Check dependencies
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        log_error "This script requires either curl or wget to download files"
        log_error "Please install one of these tools and try again"
        exit 1
    fi
    
    # Detect platform
    platform=$(detect_platform)
    log_info "Detected platform: $platform"
    
    # Get latest version
    version=$(get_latest_version)
    log_info "Latest version: $version"
    
    # Download binary
    temp_file=$(download_binary "$version" "$platform")
    
    # Install binary
    install_path=$(install_binary "$temp_file")
    
    # Cleanup
    cleanup "$temp_file"
    
    # Verify installation
    if [ -x "$install_path" ]; then
        log_success "${BINARY_NAME} installed successfully to: $install_path"
        
        # Check PATH
        check_path "$install_path"
        
        # Show version
        if command -v "$BINARY_NAME" >/dev/null 2>&1; then
            log_info "Installed version: $("$BINARY_NAME" --version 2>/dev/null || echo "version check failed")"
        else
            log_info "Binary installed but not found in PATH. Use full path: $install_path"
        fi
        
        log_success "Installation complete! 🎉"
        log_info "Run '${BINARY_NAME} --help' to get started"
    else
        log_error "Installation failed - binary not executable"
        exit 1
    fi
}

# Handle command line arguments
case "${1:-}" in
    -h|--help)
        echo "sortpath installer script"
        echo ""
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  -h, --help     Show this help message"
        echo "  -v, --verbose  Enable verbose output"
        echo ""
        echo "Environment variables:"
        echo "  INSTALL_DIR    Override installation directory (default: $INSTALL_DIR)"
        echo ""
        echo "Supported platforms:"
        echo "  - Linux (amd64, arm64)"
        echo "  - macOS (amd64, arm64)"
        echo "  - Windows (amd64)"
        exit 0
        ;;
    -v|--verbose)
        set -x
        ;;
esac

# Override install directory if specified
if [ -n "${INSTALL_DIR:-}" ]; then
    INSTALL_DIR="$INSTALL_DIR"
fi

# Run main installation
main 