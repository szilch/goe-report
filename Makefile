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

.PHONY: all build clean run docker-build docker-run release


docker-build:
	@echo "Building Docker image $(BINARY_NAME)-cron..."
	@docker build -f docker/Dockerfile -t $(BINARY_NAME)-cron .

release:
	@CURRENT_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	NEXT_VERSION=$$(echo $$CURRENT_TAG | sed 's/^v//' | awk -F. '{printf "v%d.%d.0", $$1, $$2+1}'); \
	read -p "Enter version [$$NEXT_VERSION]: " VERSION; \
	VERSION=$${VERSION:-$$NEXT_VERSION}; \
	if [ "$${VERSION#v}" = "$$VERSION" ]; then VERSION="v$$VERSION"; fi; \
	echo "Releasing version $$VERSION..."; \
	git tag -a $$VERSION -m "$$VERSION"; \
	git push origin $$VERSION
