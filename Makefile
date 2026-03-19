BINARY_NAME=echarge-report
BUILD_DIR=bin
VERSION=$(shell git describe --tags --exact-match 2>/dev/null || echo "dev")
LDFLAGS=-s -w -X echarge-report/pkg/version.Version=$(VERSION)

all: build

build:
	@echo "Building $(BINARY_NAME) (version: $(VERSION))..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -f echarge-report
	@echo "Clean complete."

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: all build clean run docker-build docker-run

docker-build:
	@echo "Building Docker image $(BINARY_NAME)-cron..."
	@docker build -f docker/Dockerfile -t $(BINARY_NAME)-cron .
