# PDF Reader SDK Makefile

.PHONY: help build test clean run-example install deps fmt vet lint benchmark coverage

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the SDK"
	@echo "  test         - Run tests"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  benchmark    - Run benchmark tests"
	@echo "  coverage     - Run tests with coverage report"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Run golint (requires golint to be installed)"
	@echo "  deps         - Download dependencies"
	@echo "  clean        - Clean build artifacts"
	@echo "  run-example  - Run the example (requires PDF_FILE environment variable)"
	@echo "  install      - Install the SDK locally"

# Build the SDK
build:
	@echo "Building PDF Reader SDK..."
	go build -v ./...

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -race ./...

# Run benchmark tests
benchmark:
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run golint (requires golint to be installed)
lint:
	@echo "Running golint..."
	@command -v golint >/dev/null 2>&1 || { echo "golint not installed. Install with: go install golang.org/x/lint/golint@latest"; exit 1; }
	golint ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	go clean ./...
	rm -f coverage.out coverage.html

# Run the example
run-example:
	@if [ -z "$(PDF_FILE)" ]; then \
		echo "Error: PDF_FILE environment variable is required"; \
		echo "Usage: make run-example PDF_FILE=/path/to/your/file.pdf"; \
		exit 1; \
	fi
	@echo "Running example with PDF file: $(PDF_FILE)"
	cd examples && go run main.go "$(PDF_FILE)"

# Install the SDK locally
install:
	@echo "Installing PDF Reader SDK..."
	go install ./...

# Run all quality checks
check: fmt vet test
	@echo "All quality checks passed!"

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	go mod download
	@echo "Installing development tools..."
	go install golang.org/x/lint/golint@latest
	@echo "Development setup complete!"

# Create a sample test PDF (requires a PDF file to copy)
create-test-pdf:
	@if [ -z "$(SOURCE_PDF)" ]; then \
		echo "Error: SOURCE_PDF environment variable is required"; \
		echo "Usage: make create-test-pdf SOURCE_PDF=/path/to/source.pdf"; \
		exit 1; \
	fi
	@echo "Creating test PDF from $(SOURCE_PDF)..."
	cp "$(SOURCE_PDF)" examples/sample.pdf
	@echo "Test PDF created: examples/sample.pdf"

# Run example with sample PDF
run-sample:
	@if [ ! -f "examples/sample.pdf" ]; then \
		echo "Error: examples/sample.pdf not found"; \
		echo "Create it first with: make create-test-pdf SOURCE_PDF=/path/to/your.pdf"; \
		exit 1; \
	fi
	@echo "Running example with sample PDF..."
	cd examples && go run main.go sample.pdf

# Show project statistics
stats:
	@echo "Project Statistics:"
	@echo "==================="
	@echo "Go files:"
	@find . -name "*.go" -not -path "./vendor/*" | wc -l
	@echo "Lines of code:"
	@find . -name "*.go" -not -path "./vendor/*" -exec wc -l {} + | tail -1
	@echo "Test files:"
	@find . -name "*_test.go" -not -path "./vendor/*" | wc -l
	@echo "Dependencies:"
	@go list -m all | wc -l

# Release build (with optimizations)
release:
	@echo "Building release version..."
	go build -ldflags="-s -w" -v ./...
	@echo "Release build complete!"
