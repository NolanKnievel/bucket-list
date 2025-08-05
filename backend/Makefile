# Makefile for Collaborative Bucket List Backend

# Variables
BINARY_NAME=collaborative-bucket-list
MAIN_PATH=./cmd/main.go
BUILD_DIR=./build
DOCKER_IMAGE=collaborative-bucket-list-backend
DOCKER_TAG=latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
BUILD_FLAGS=-ldflags="-w -s"
PROD_BUILD_FLAGS=-ldflags="-w -s" -a -installsuffix cgo

.PHONY: all build build-prod clean test deps run docker-build docker-run help

# Default target
all: clean deps test build

# Build for development
build:
	@echo "Building for development..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build for production
build-prod:
	@echo "Building for production..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(PROD_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build for production with version info
build-prod-with-version:
	@echo "Building for production with version info..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(PROD_BUILD_FLAGS) \
		-ldflags="-w -s -X main.Version=$(shell git describe --tags --always --dirty) \
		-X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S') \
		-X main.CommitHash=$(shell git rev-parse --short HEAD)" \
		-o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Run the application
run:
	@echo "Running application..."
	$(GOCMD) run $(MAIN_PATH)

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

# Build production Docker image
docker-build-prod:
	@echo "Building production Docker image..."
	docker build --target production -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

# Run production Docker container
docker-run-prod:
	@echo "Running production Docker container..."
	docker run -p 8080:8080 -p 9090:9090 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

# Build for multiple platforms
build-multi-platform:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(PROD_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(PROD_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(PROD_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

# Security scan
security-scan:
	@echo "Running security scan..."
	gosec ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Help
help:
	@echo "Available targets:"
	@echo "  all                    - Clean, download deps, test, and build"
	@echo "  build                  - Build for development"
	@echo "  build-prod             - Build for production (Linux AMD64)"
	@echo "  build-prod-with-version- Build for production with version info"
	@echo "  build-multi-platform   - Build for multiple platforms"
	@echo "  clean                  - Clean build artifacts"
	@echo "  test                   - Run tests"
	@echo "  test-coverage          - Run tests with coverage report"
	@echo "  deps                   - Download and tidy dependencies"
	@echo "  run                    - Run the application"
	@echo "  docker-build           - Build Docker image"
	@echo "  docker-build-prod      - Build production Docker image"
	@echo "  docker-run             - Run Docker container"
	@echo "  docker-run-prod        - Run production Docker container"
	@echo "  security-scan          - Run security scan"
	@echo "  lint                   - Lint code"
	@echo "  help                   - Show this help message"