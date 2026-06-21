# Makefile for Gohabit TUI

BINARY_NAME=gohabit
INSTALL_DIR=$(HOME)/.local/bin
VERSION=0.0.1
DIST_DIR=dist

.PHONY: all build install uninstall clean release help

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME)

install:
	@chmod +x install.sh
	@./install.sh

uninstall:
	@echo "Removing $(BINARY_NAME) from $(INSTALL_DIR)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstalled successfully."

clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(DIST_DIR)

release:
	@echo "Creating release binaries in $(DIST_DIR)..."
	@mkdir -p $(DIST_DIR)
	
	# Linux amd64
	@echo "Building Linux amd64..."
	@GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)
	@tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_linux_amd64.tar.gz -C $(DIST_DIR) $(BINARY_NAME)
	
	# Linux arm64
	@echo "Building Linux arm64..."
	@GOOS=linux GOARCH=arm64 go build -o $(DIST_DIR)/$(BINARY_NAME)
	@tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_linux_arm64.tar.gz -C $(DIST_DIR) $(BINARY_NAME)
	
	# macOS amd64
	@echo "Building macOS amd64..."
	@GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/$(BINARY_NAME)
	@tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_amd64.tar.gz -C $(DIST_DIR) $(BINARY_NAME)
	
	# macOS arm64
	@echo "Building macOS arm64..."
	@GOOS=darwin GOARCH=arm64 go build -o $(DIST_DIR)/$(BINARY_NAME)
	@tar -czf $(DIST_DIR)/$(BINARY_NAME)_$(VERSION)_darwin_arm64.tar.gz -C $(DIST_DIR) $(BINARY_NAME)
	
	# Clean temporary binaries in dist
	@rm -f $(DIST_DIR)/$(BINARY_NAME)
	@echo "All release binaries successfully packaged in $(DIST_DIR)/"

help:
	@echo "Usage:"
	@echo "  make          - Build the gohabit binary locally"
	@echo "  make build    - Build the gohabit binary locally"
	@echo "  make install  - Build and install gohabit to $(INSTALL_DIR)"
	@echo "  make uninstall- Remove gohabit from $(INSTALL_DIR)"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make release  - Compile and package release binaries for Linux/macOS in $(DIST_DIR)"
