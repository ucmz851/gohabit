# Makefile for Gohabit TUI

BINARY_NAME=gohabit
INSTALL_DIR=$(HOME)/.local/bin

.PHONY: all build install uninstall clean help

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

help:
	@echo "Usage:"
	@echo "  make          - Build the gohabit binary locally"
	@echo "  make build    - Build the gohabit binary locally"
	@echo "  make install  - Build and install gohabit to $(INSTALL_DIR)"
	@echo "  make uninstall- Remove gohabit from $(INSTALL_DIR)"
	@echo "  make clean    - Remove build artifacts"
