# Variables
GO=go
BUILD_DIR=build
BINARY_NAME=j2z
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)

# Targets
all: build

build:
	$(GO) build -o $(BINARY_PATH) cmd/j2z/main.go

clean:
	rm -rf $(BUILD_DIR)

.PHONY: all build clean

# End of file
