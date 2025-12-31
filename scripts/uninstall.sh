#!/bin/bash

# Oh Man! Uninstall Script

set -e

BINARY_NAME="ohman"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/ohman"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Remove binary
remove_binary() {
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        info "Removing $INSTALL_DIR/$BINARY_NAME"
        if [ -w "$INSTALL_DIR" ]; then
            rm "$INSTALL_DIR/$BINARY_NAME"
        else
            sudo rm "$INSTALL_DIR/$BINARY_NAME"
        fi
    else
        warn "Not found: $INSTALL_DIR/$BINARY_NAME"
    fi
}

# Remove config files
remove_config() {
    echo ""
    info "Do you want to remove the config files ($CONFIG_DIR)? [y/N]"
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        if [ -d "$CONFIG_DIR" ]; then
            rm -rf "$CONFIG_DIR"
            info "Config files removed"
        fi
    else
        info "Keeping config files"
    fi
}

# Remove shell hook
remove_hook() {
    echo ""
    info "Do you want to remove the shell hook? [Y/n]"
    read -r response
    
    if [[ "$response" =~ ^[Nn]$ ]]; then
        return
    fi

    SHELL_NAME=$(basename "$SHELL")
    
    case $SHELL_NAME in
        zsh)
            HOOK_FILE="$HOME/.zshrc"
            ;;
        bash)
            HOOK_FILE="$HOME/.bashrc"
            ;;
        *)
            return
            ;;
    esac

    if [ -f "$HOOK_FILE" ]; then
        # Create backup
        cp "$HOOK_FILE" "$HOOK_FILE.ohman.bak"
        
        # Remove hook-related lines
        sed -i.bak '/# Oh Man!/d' "$HOOK_FILE"
        sed -i.bak '/ohman_precmd/d' "$HOOK_FILE"
        sed -i.bak '/ohman_prompt_command/d' "$HOOK_FILE"
        sed -i.bak '/precmd_functions+=(ohman_precmd)/d' "$HOOK_FILE"
        
        rm -f "$HOOK_FILE.bak"
        
        info "Hook removed from $HOOK_FILE"
        info "Backup saved at $HOOK_FILE.ohman.bak"
    fi
}

# Clean up temp files
cleanup_temp() {
    rm -f /tmp/.ohman_last_failed_* 2>/dev/null || true
}

main() {
    echo "================================"
    echo "  Oh Man! Uninstaller"
    echo "================================"
    echo ""
    
    remove_binary
    remove_config
    remove_hook
    cleanup_temp
    
    echo ""
    info "Oh Man! has been uninstalled!"
    echo ""
}

main
