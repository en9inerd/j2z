# Variables
GO=go
BUILD_DIR=build
DIST_DIR=dist
BINARY_NAME=$(shell basename $(PWD))
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)

# Targets
all: build

build:
	$(GO) build -o $(BINARY_PATH) ./cmd/j2z/

build-prod:
	bash scripts/build.sh

test:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)

.PHONY: all build build-prod test vet clean
