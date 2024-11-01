# Variables
GO=go
BUILD_DIR=build
BINARY_NAME=j2z
BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)
CGO_ENABLED=0

# Targets
all: build

build:
	$(GO) build -o $(BINARY_PATH) ./src

build-prod:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -gcflags="all=-l -B" -trimpath -ldflags="-s -w" -o $(BINARY_PATH) ./src

clean:
	rm -rf $(BUILD_DIR)

.PHONY: all build clean

# End of file
