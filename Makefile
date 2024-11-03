# Variables
GO=go
BUILD_DIR=build
DIST_DIR=dist
BINARY_NAME=$(shell basename $(PWD))
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)

# Targets
all: build

build:
	$(GO) build -o $(BINARY_PATH) ./src

build-prod:
	bash scripts/build.sh

clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)

.PHONY: all build clean

# End of file
