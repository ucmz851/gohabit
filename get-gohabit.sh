#!/bin/sh

# get-gohabit.sh - Standalone installer for Gohabit TUI
# Can be run via: curl -sSfL https://raw.githubusercontent.com/ucmz851/gohabit/main/get-gohabit.sh | sh

set -e

REPO="ucmz851/gohabit"
BINARY_NAME="gohabit"
INSTALL_DIR="$HOME/.local/bin"

mkdir -p "$INSTALL_DIR"

# Check if Go is installed
if command -v go >/dev/null 2>&1; then
    echo "Go is installed. Building from source..."
    TEMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TEMP_DIR"' EXIT
    
    echo "Cloning repository..."
    git clone --depth 1 "https://github.com/$REPO.git" "$TEMP_DIR"
    
    echo "Compiling gohabit..."
    (cd "$TEMP_DIR" && go build -o "$BINARY_NAME")
    
    cp "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    echo "Successfully built and copied binary to $INSTALL_DIR."
else
    echo "Go is not installed. Downloading pre-compiled binary..."
    
    # Detect OS
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$OS" in
        darwin)  OS="darwin" ;;
        linux)   OS="linux" ;;
        freebsd) OS="freebsd" ;;
        *) echo "Unsupported OS: $OS"; exit 1 ;;
    esac
    
    # Detect Architecture
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) echo "Unsupported Architecture: $ARCH"; exit 1 ;;
    esac
    
    echo "Detected platform: ${OS}/${ARCH}"
    
    # Get latest release tag from GitHub API (cross-platform sed parse)
    echo "Fetching latest release version..."
    LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"tag_name":\s*"([^"]+)".*/\1/')
    
    if [ -z "$LATEST_RELEASE" ]; then
        # Fallback if API fails
        LATEST_RELEASE="v1.0.0"
    fi
    
    echo "Latest release is $LATEST_RELEASE"
    
    # Download URL (Assuming standard release asset naming: gohabit_1.0.0_linux_amd64.tar.gz)
    VERSION=$(echo "$LATEST_RELEASE" | sed 's/^v//')
    
    TARBALL="gohabit_${VERSION}_${OS}_${ARCH}.tar.gz"
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_RELEASE/$TARBALL"
    
    TEMP_DIR=$(mktemp -d)
    trap 'rm -rf "$TEMP_DIR"' EXIT
    
    echo "Downloading from $DOWNLOAD_URL..."
    if command -v curl >/dev/null 2>&1; then
        curl -sSfL -o "$TEMP_DIR/$TARBALL" "$DOWNLOAD_URL"
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O "$TEMP_DIR/$TARBALL" "$DOWNLOAD_URL"
    else
        echo "Error: Neither curl nor wget is installed."
        exit 1
    fi
    
    echo "Extracting binary..."
    tar -xzf "$TEMP_DIR/$TARBALL" -C "$TEMP_DIR"
    
    # Copy binary to destination
    cp "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    echo "Successfully downloaded and installed binary to $INSTALL_DIR."
fi

# Ensure INSTALL_DIR is in PATH
RESOLVED_INSTALL_DIR=$(cd "$INSTALL_DIR" && pwd)
case ":$PATH:" in
    *":$RESOLVED_INSTALL_DIR:"*)
        echo "Great! $INSTALL_DIR is already in your PATH."
        ;;
    *)
        echo "Warning: $INSTALL_DIR is not in your PATH."
        echo "Attempting to add it to your shell profile..."
        
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
            SHELL_CONF="$HOME/.profile"
            echo "" >> "$SHELL_CONF"
            echo "# Added by gohabit installer" >> "$SHELL_CONF"
            echo "export PATH=\"\$PATH:$RESOLVED_INSTALL_DIR\"" >> "$SHELL_CONF"
            echo "Created and updated $SHELL_CONF. Please run: source $SHELL_CONF"
        fi
        ;;
esac

echo "Gohabit is ready! Run 'gohabit' to start tracking your habits."
