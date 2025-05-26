# Awesome Tools

A collection of useful Go utilities and SDKs for various development tasks.

## Overview

This repository contains a set of Go tools and libraries designed to simplify common development workflows. Currently includes:

- **PDF Reader SDK** - A powerful Go library for PDF text extraction and processing

## Project Structure

```
awesome-tools/
├── pdf-reader/          # PDF processing SDK
│   ├── reader.go        # Core PDF reader implementation
│   ├── utils.go         # Utility functions
│   ├── examples/        # Usage examples
│   └── README.md        # Detailed PDF reader documentation
├── go.mod               # Go module definition
├── Makefile             # Build automation
└── README.md            # This file
```

## Quick Start

### Prerequisites

- Go 1.24.1 or later

### Installation

Clone the repository:

```bash
git clone <repository-url>
cd awesome-tools
```

Install dependencies:

```bash
make deps
```

### Building

Build all tools:

```bash
make build
```

### Testing

Run all tests:

```bash
make test
```

## Available Tools

### PDF Reader SDK

A comprehensive Go SDK for reading and processing PDF documents.

**Key Features:**
- Extract text from PDF files
- Search text within PDFs
- Get document metadata
- Process PDFs from URLs
- Batch processing support

**Quick Example:**

```go
package main

import (
    "fmt"
    "log"
    
    pdfreader "awesome-tools/pdf-reader"
)

func main() {
    text, err := pdfreader.ExtractTextFromFile("document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Extracted text:", text)
}
```

For detailed documentation, see [pdf-reader/README.md](pdf-reader/README.md).

### Running Examples

To run the PDF reader example:

```bash
make run-example PDF_FILE=/path/to/your/file.pdf
```

## Development

### Available Make Commands

- `make build` - Build all tools
- `make test` - Run tests
- `make fmt` - Format code
- `make vet` - Run go vet
- `make lint` - Run golint (requires golint installation)
- `make deps` - Download and tidy dependencies
- `make run-example` - Run PDF reader example (requires PDF_FILE env var)
- `make install` - Install tools locally

### Code Quality

This project follows Go best practices:

- Code formatting with `go fmt`
- Static analysis with `go vet`
- Linting with `golint`
- Comprehensive testing

Run all quality checks:

```bash
make fmt && make vet && make lint && make test
```

## Dependencies

- `github.com/ledongthuc/pdf` - PDF parsing and text extraction

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and quality checks
5. Submit a pull request

## Support

For issues and questions, please open an issue in the repository.
