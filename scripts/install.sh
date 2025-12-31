#!/bin/bash

# Oh Man! Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/liliang-cn/ohman/main/scripts/install.sh | bash

set -e

REPO="liliang-cn/ohman"
BINARY_NAME="ohman"
INSTALL_DIR="/usr/local/bin"

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print info
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case $OS in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        *)
            error "Unsupported operating system: $OS"
            ;;
    esac

    case $ARCH in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    info "Detected platform: $PLATFORM"
}

# Get latest version
get_latest_version() {
    LATEST_VERSION=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$LATEST_VERSION" ]; then
        error "Unable to get latest version info"
    fi
    info "Latest version: $LATEST_VERSION"
}

# Download and install
install() {
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${BINARY_NAME}-${PLATFORM}"
    
    info "Downloading: $DOWNLOAD_URL"
    
    TMP_FILE=$(mktemp)
    if ! curl -sL "$DOWNLOAD_URL" -o "$TMP_FILE"; then
        error "Download failed"
    fi

    chmod +x "$TMP_FILE"

    info "Installing to $INSTALL_DIR/$BINARY_NAME"
    
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    else
        sudo mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    fi

    info "Installation successful!"
}

# Install shell hook
install_hook() {
    echo ""
    info "Do you want to install the shell hook for automatic failed command detection? [Y/n]"
    read -r response
    
    if [[ "$response" =~ ^[Nn]$ ]]; then
        warn "Skipping hook installation. Failed command diagnosis feature may be limited."
        return
    fi

    SHELL_NAME=$(basename "$SHELL")
    
    case $SHELL_NAME in
        zsh)
            HOOK_FILE="$HOME/.zshrc"
            HOOK_SCRIPT='
# Oh Man! Failed command recording hook
ohman_precmd() {
    local exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        echo "$exit_code|$(fc -ln -1)|$(date +%s)" > /tmp/.ohman_last_failed_$$
    fi
}
precmd_functions+=(ohman_precmd)'
            ;;
        bash)
            HOOK_FILE="$HOME/.bashrc"
            HOOK_SCRIPT='
# Oh Man! Failed command recording hook
ohman_prompt_command() {
    local exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        echo "$exit_code|$(history 1 | sed '"'"'s/^[ ]*[0-9]*[ ]*//'"'"')|$(date +%s)" > /tmp/.ohman_last_failed_$$
    fi
}
PROMPT_COMMAND="ohman_prompt_command${PROMPT_COMMAND:+; $PROMPT_COMMAND}"'
            ;;
        *)
            warn "Unsupported shell: $SHELL_NAME, skipping hook installation"
            return
            ;;
    esac

    # Check if already installed
    if grep -q "ohman_precmd\|ohman_prompt_command" "$HOOK_FILE" 2>/dev/null; then
        info "Hook already exists in $HOOK_FILE"
        return
    fi

    echo "$HOOK_SCRIPT" >> "$HOOK_FILE"
    info "Hook added to $HOOK_FILE"
    info "Please run 'source $HOOK_FILE' or reopen your terminal for changes to take effect"
}

# Display usage instructions
show_usage() {
    echo ""
    echo "============================================"
    echo "  Oh Man! Installation Complete! 🎉"
    echo "============================================"
    echo ""
    echo "Quick Start:"
    echo "  1. Configure LLM:  ohman config"
    echo "  2. Start using:    ohman grep \"How to search recursively?\""
    echo ""
    echo "More info: ohman --help"
    echo ""
}

# Main flow
main() {
    echo "================================"
    echo "  Oh Man! Installer"
    echo "================================"
    echo ""
    
    detect_platform
    get_latest_version
    install
    install_hook
    show_usage
}

main
