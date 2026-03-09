BINARY_NAME=goe-report
BUILD_DIR=bin

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -f goe-report
	@echo "Clean complete."

run: build
	@./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: all build clean run docker-build docker-run

docker-build:
	@echo "Building Docker image $(BINARY_NAME)-cron..."
	@docker build -f docker/Dockerfile -t $(BINARY_NAME)-cron .
