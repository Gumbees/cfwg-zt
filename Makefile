# Makefile for Cloudflare Zero Trust WireGuard Manager for UDM-Pro

BINARY_NAME=cfwg-zt
VERSION=1.0.0
BUILD_DIR=build
PLATFORMS=linux/arm64 linux/amd64

.PHONY: all build clean test package

all: clean build test

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output=$(BUILD_DIR)/$(BINARY_NAME)_$${os}_$${arch}; \
		echo "Building for $${os}/$${arch}..."; \
		GOOS=$$os GOARCH=$$arch go build -o $$output ./cmd/$(BINARY_NAME); \
	done

udm-pro:
	@echo "Building for UDM-Pro (linux/arm64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	go test -v ./...

package: clean udm-pro
	@echo "Creating release package..."
	@mkdir -p $(BUILD_DIR)/release
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(BUILD_DIR)/release/
	@cp -r install $(BUILD_DIR)/release/
	@cp README.md $(BUILD_DIR)/release/
	@cd $(BUILD_DIR) && tar -czvf $(BINARY_NAME)-$(VERSION).tar.gz -C release .
	@echo "Package created: $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION).tar.gz"
