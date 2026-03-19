.PHONY: help build test clean run-example install deps fmt vet lint benchmark coverage status-updater-build status-updater-run status-updater-test
build:
	@echo "Building awesome-tools SDK..."
	go build -v ./...

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Check pdf-reader test coverage
pdfreader-test-cov:
	@echo "Checking pdf-reader package test coverage..."
	cd pdf-reader && go test -v -cover

# Check dttm test coverage
dttm-test-cov:
	@echo "Checking dttm package test coverage..."
	cd dttm && go test -v -cover

# Check go-struct-utils test coverage
gostructutils-test-cov:
	@echo "Checking go-struct-utils package test coverage..."
	cd go-struct-utils && go test -v -cover

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

# Run the example
run-example:
	@if [ -z "$(PDF_FILE)" ]; then \
		echo "Error: PDF_FILE environment variable is required"; \
		echo "Usage: make run-example -file-loc=/path/to/your/file.pdf"; \
		exit 1; \
	fi
	@echo "Running example with PDF file: $(PDF_FILE)"
	cd pdf-reader/examples && go run main.go -file-loc="$(PDF_FILE)"

# Install the SDK locally
install:
	@echo "Installing awesome-tools SDK..."
	go install ./...

# Build the status-updater binary
status-updater-build:
	@echo "Building status-updater..."
	go build -o bin/status-updater ./status-updater

# Run status-updater (requires FROM and TO, e.g. make status-updater-run FROM=2024-01-01 TO=2024-01-07)
status-updater-run:
	@if [ -z "$(FROM)" ] || [ -z "$(TO)" ]; then \
		echo "Usage: make status-updater-run FROM=YYYY-MM-DD TO=YYYY-MM-DD [OUTPUT=file.md]"; \
		exit 1; \
	fi
	@if [ -n "$(OUTPUT)" ]; then \
		go run ./status-updater --from $(FROM) --to $(TO) --output $(OUTPUT); \
	else \
		go run ./status-updater --from $(FROM) --to $(TO); \
	fi

# Run status-updater tests
status-updater-test:
	@echo "Running status-updater tests..."
	go test -v -cover ./status-updater/...

# generates commit message and commits
commit:
	@echo "Downloading git-commit script..."
	@curl -s -o git-commit.gpt https://raw.githubusercontent.com/HiteshRepo/gpt-script-tool/main/tools/github/git-commit.gpt
	@echo "Commiting changes..."
	@gptscript --disable-cache git-commit.gpt
	@echo "Cleaning up downloaded script..."
	@rm -f git-commit.gpt
