# Awesome Tools

A collection of useful Go utilities and SDKs for various development tasks.

## Overview

This repository contains a set of Go tools and libraries designed to simplify common development workflows. Currently includes:

- **PDF Reader SDK** - A powerful Go library for PDF text extraction and processing
- **DTTM** - Date/time utilities for flexible parsing and formatting with multiple format support
- **Go Struct Utils** - Utilities for converting Go structs to maps using different approaches

## Project Structure

```
awesome-tools/
├── pdf-reader/          # PDF processing SDK
│   ├── reader.go        # Core PDF reader implementation
│   ├── utils.go         # Utility functions
│   ├── examples/        # Usage examples
│   └── README.md        # Detailed PDF reader documentation
├── dttm/                # Date/time utilities
│   ├── dttm.go          # Core date/time functions
│   └── README.md        # DTTM documentation
├── go-struct-utils/     # Struct to map conversion utilities
│   ├── utils.go         # Conversion functions
│   ├── utils_test.go    # Comprehensive tests
│   └── README.md        # Go struct utils documentation
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

### DTTM - Date/Time Utilities

A flexible Go package for parsing and formatting time strings in various formats.

**Key Features:**
- Multiple time format support (RFC3339, human-readable, custom formats)
- Automatic format detection during parsing
- Time extraction from text containing other content
- UTC conversion for all parsed times
- Format conversion between different representations

**Quick Example:**

```go
package main

import (
    "fmt"
    "log"
    
    "awesome-tools/dttm"
)

func main() {
    // Parse time string (auto-detects format)
    t, err := dttm.ParseTime("25-Dec-2023_15:30:45")
    if err != nil {
        log.Fatal(err)
    }
    
    // Convert to different formats
    standard := dttm.FormatTo(t, dttm.Standard)
    dateOnly := dttm.FormatTo(t, dttm.DateOnly)
    
    fmt.Printf("Standard: %s\n", standard)
    fmt.Printf("Date Only: %s\n", dateOnly)
}
```

For detailed documentation, see [dttm/README.md](dttm/README.md).

### Go Struct Utils

Utilities for converting Go structs to maps using different approaches and strategies.

**Key Features:**
- Three different conversion methods (JSON-based, basic reflection, advanced reflection)
- Full JSON tag support with proper handling of exclusions
- Handles nested structs, pointers, and edge cases
- Type preservation options
- Comprehensive test coverage

**Quick Example:**

```go
package main

import (
    "fmt"
    "log"
    
    gostructutils "awesome-tools/go-struct-utils"
)

type Person struct {
    Name     string `json:"name"`
    Age      int    `json:"age"`
    Password string `json:"-"`  // Excluded
}

func main() {
    person := Person{Name: "John", Age: 30, Password: "secret"}
    
    // JSON-based conversion (respects all JSON tags)
    result, err := gostructutils.StructToMapJSON(person)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Result: %+v\n", result)
    // Output: map[age:30 name:John]
}
```

For detailed documentation, see [go-struct-utils/README.md](go-struct-utils/README.md).

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
- `github.com/pkg/errors` - Enhanced error handling (used by DTTM package)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and quality checks
5. Submit a pull request

## Support

For issues and questions, please open an issue in the repository.
