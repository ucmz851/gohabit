#!/bin/sh

# install.sh - Installer for Gohabit TUI
# Compiles and copies the binary to ~/.local/bin, ensuring the directory is in the system PATH.

set -e

BINARY_NAME="gohabit"
INSTALL_DIR="$HOME/.local/bin"

echo "Building $BINARY_NAME..."
go build -o "$BINARY_NAME"

echo "Installing $BINARY_NAME to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
cp "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
echo "Successfully copied binary to $INSTALL_DIR."

# Check if INSTALL_DIR is in PATH
# Fully resolve paths to avoid home tilde mismatch
RESOLVED_INSTALL_DIR=$(cd "$INSTALL_DIR" && pwd)

case ":$PATH:" in
    *":$RESOLVED_INSTALL_DIR:"*)
        echo "Great! $INSTALL_DIR is already in your PATH."
        ;;
    *)
        echo "Warning: $INSTALL_DIR is not in your PATH."
        echo "Attempting to add it to your shell profile..."
        
        # Detect shell configuration file
        SHELL_CONF=""
        case "$SHELL" in
            */bash)
                if [ -f "$HOME/.bashrc" ]; then
                    SHELL_CONF="$HOME/.bashrc"
                elif [ -f "$HOME/.bash_profile" ]; then
                    SHELL_CONF="$HOME/.bash_profile"
                else
                    SHELL_CONF="$HOME/.profile"
                fi
                ;;
            */zsh)
                SHELL_CONF="$HOME/.zshrc"
                ;;
            *)
                SHELL_CONF="$HOME/.profile"
                ;;
        esac

        if [ -n "$SHELL_CONF" ] && [ -f "$SHELL_CONF" ]; then
            # Check if it was already appended to avoid duplicates
            if grep -Fq "$RESOLVED_INSTALL_DIR" "$SHELL_CONF"; then
                echo "Path configuration already exists in $SHELL_CONF."
            else
                echo "" >> "$SHELL_CONF"
                echo "# Added by gohabit installer" >> "$SHELL_CONF"
                echo "export PATH=\"\$PATH:$RESOLVED_INSTALL_DIR\"" >> "$SHELL_CONF"
                echo "Added 'export PATH=\"\$PATH:$RESOLVED_INSTALL_DIR\"' to $SHELL_CONF"
                echo "Please run: source $SHELL_CONF (or restart your terminal) to apply changes."
            fi
        else
            # Create ~/.profile if no shell conf exists
            SHELL_CONF="$HOME/.profile"
            echo "" >> "$SHELL_CONF"
            echo "# Added by gohabit installer" >> "$SHELL_CONF"
            echo "export PATH=\"\$PATH:$RESOLVED_INSTALL_DIR\"" >> "$SHELL_CONF"
            echo "Created and updated $SHELL_CONF. Please run: source $SHELL_CONF"
        fi
        ;;
esac
