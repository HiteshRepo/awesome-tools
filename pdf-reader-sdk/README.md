# PDF Reader SDK

A powerful and easy-to-use Golang SDK for reading PDF documents and extracting text content. This SDK provides a clean API for PDF text extraction with support for various options and utilities.

## Features

- 📄 Extract text from PDF files
- 🔍 Search for specific text within PDFs
- 📊 Get document metadata and information
- 🎯 Extract text from specific pages or page ranges
- 🌐 Download and process PDFs from URLs
- 📁 Batch processing of multiple PDF files
- ✅ PDF file validation
- 🛠 Flexible text extraction options

## Installation

```bash
go get github.com/your-username/pdf-reader-sdk
```

Or if using this locally:

```bash
go mod init your-project
go get ./pdf-reader-sdk
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    pdfreader "pdf-reader-sdk"
)

func main() {
    // Simple text extraction
    text, err := pdfreader.ExtractTextFromFile("document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Extracted text:", text)
}
```

## API Reference

### Core Types

#### PDFReader

The main struct for reading PDF documents.

```go
type PDFReader struct {
    // Internal fields
}
```

#### TextExtractOptions

Options for customizing text extraction.

```go
type TextExtractOptions struct {
    PageRange          *PageRange // Specify which pages to extract
    PreserveFormatting bool       // Attempt to maintain text formatting
    JoinLines          bool       // Join text lines with spaces
}
```

#### PageRange

Defines a range of pages to extract.

```go
type PageRange struct {
    Start int // 1-based page number
    End   int // 1-based page number, -1 means last page
}
```

#### DocumentInfo

Contains metadata about the PDF document.

```go
type DocumentInfo struct {
    Title        string
    Author       string
    Subject      string
    Creator      string
    Producer     string
    CreationDate string
    ModDate      string
    PageCount    int
}
```

### Core Functions

#### NewPDFReader

Creates a new PDF reader from a file path.

```go
func NewPDFReader(filePath string) (*PDFReader, error)
```

**Example:**
```go
reader, err := pdfreader.NewPDFReader("document.pdf")
if err != nil {
    log.Fatal(err)
}
defer reader.Close()
```

#### NewPDFReaderFromReader

Creates a new PDF reader from an io.ReaderAt.

```go
func NewPDFReaderFromReader(reader io.ReaderAt, size int64) (*PDFReader, error)
```

### Text Extraction Methods

#### ExtractText

Extracts text from the PDF document with the given options.

```go
func (pr *PDFReader) ExtractText(options *TextExtractOptions) (string, error)
```

**Example:**
```go
options := &pdfreader.TextExtractOptions{
    PageRange: &pdfreader.PageRange{Start: 1, End: 5},
    JoinLines: true,
}

text, err := reader.ExtractText(options)
if err != nil {
    log.Fatal(err)
}
```

#### ExtractTextFromPage

Extracts text from a specific page (1-based).

```go
func (pr *PDFReader) ExtractTextFromPage(pageNum int) (string, error)
```

**Example:**
```go
pageText, err := reader.ExtractTextFromPage(1)
if err != nil {
    log.Fatal(err)
}
```

### Utility Functions

#### ExtractTextFromFile

Convenience function to extract text from a PDF file.

```go
func ExtractTextFromFile(filePath string) (string, error)
```

#### ExtractTextFromFileWithOptions

Extracts text from a PDF file with custom options.

```go
func ExtractTextFromFileWithOptions(filePath string, options *TextExtractOptions) (string, error)
```

#### ExtractTextFromBytes

Extracts text from PDF data in memory.

```go
func ExtractTextFromBytes(data []byte) (string, error)
```

#### ExtractTextFromURL

Downloads a PDF from a URL and extracts text.

```go
func ExtractTextFromURL(url string) (string, error)
```

**Example:**
```go
text, err := pdfreader.ExtractTextFromURL("https://example.com/document.pdf")
if err != nil {
    log.Fatal(err)
}
```

#### BatchExtractText

Extracts text from multiple PDF files.

```go
func BatchExtractText(filePaths []string) (map[string]string, error)
```

**Example:**
```go
files := []string{"doc1.pdf", "doc2.pdf", "doc3.pdf"}
results, err := pdfreader.BatchExtractText(files)
if err != nil {
    log.Fatal(err)
}

for file, text := range results {
    fmt.Printf("File: %s\nText: %s\n\n", file, text[:100])
}
```

#### SearchTextInPDF

Searches for a specific text pattern in a PDF.

```go
func SearchTextInPDF(filePath, searchText string) ([]PageMatch, error)
```

**Example:**
```go
matches, err := pdfreader.SearchTextInPDF("document.pdf", "important")
if err != nil {
    log.Fatal(err)
}

for _, match := range matches {
    fmt.Printf("Found on page %d\n", match.PageNumber)
}
```

### Information and Validation

#### GetDocumentInfo

Returns metadata about the PDF document.

```go
func (pr *PDFReader) GetDocumentInfo() DocumentInfo
```

#### GetFileInfo

Returns basic information about a PDF file.

```go
func GetFileInfo(filePath string) (*FileInfo, error)
```

#### ValidatePDFFile

Checks if a file is a valid PDF.

```go
func ValidatePDFFile(filePath string) error
```

## Usage Examples

### Basic Text Extraction

```go
package main

import (
    "fmt"
    "log"
    
    pdfreader "pdf-reader-sdk"
)

func main() {
    text, err := pdfreader.ExtractTextFromFile("document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Extracted text:", text)
}
```

### Extract Specific Pages

```go
package main

import (
    "fmt"
    "log"
    
    pdfreader "pdf-reader-sdk"
)

func main() {
    options := &pdfreader.TextExtractOptions{
        PageRange: &pdfreader.PageRange{
            Start: 1,
            End:   3, // Extract pages 1-3
        },
        JoinLines: true,
    }
    
    text, err := pdfreader.ExtractTextFromFileWithOptions("document.pdf", options)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Extracted text from pages 1-3:", text)
}
```

### Get Document Information

```go
package main

import (
    "fmt"
    "log"
    
    pdfreader "pdf-reader-sdk"
)

func main() {
    fileInfo, err := pdfreader.GetFileInfo("document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Title: %s\n", fileInfo.DocumentInfo.Title)
    fmt.Printf("Author: %s\n", fileInfo.DocumentInfo.Author)
    fmt.Printf("Pages: %d\n", fileInfo.PageCount)
    fmt.Printf("File Size: %d bytes\n", fileInfo.FileSize)
}
```

### Search Text in PDF

```go
package main

import (
    "fmt"
    "log"
    
    pdfreader "pdf-reader-sdk"
)

func main() {
    matches, err := pdfreader.SearchTextInPDF("document.pdf", "golang")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Found 'golang' on %d pages:\n", len(matches))
    for _, match := range matches {
        fmt.Printf("- Page %d\n", match.PageNumber)
    }
}
```

### Process PDF from URL

```go
package main

import (
    "fmt"
    "log"
    
    pdfreader "pdf-reader-sdk"
)

func main() {
    text, err := pdfreader.ExtractTextFromURL("https://example.com/document.pdf")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Extracted text from URL:", text)
}
```

### Batch Processing

```go
package main

import (
    "fmt"
    "log"
    
    pdfreader "pdf-reader-sdk"
)

func main() {
    files := []string{
        "document1.pdf",
        "document2.pdf",
        "document3.pdf",
    }
    
    results, err := pdfreader.BatchExtractText(files)
    if err != nil {
        log.Fatal(err)
    }
    
    for file, text := range results {
        fmt.Printf("File: %s\n", file)
        fmt.Printf("Text length: %d characters\n", len(text))
        fmt.Printf("Preview: %s...\n\n", text[:min(100, len(text))])
    }
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

## Running the Examples

1. Navigate to the examples directory:
```bash
cd examples
```

2. Run the example with a PDF file:
```bash
go run main.go /path/to/your/document.pdf
```

## Error Handling

The SDK provides detailed error messages for common issues:

- File not found or inaccessible
- Invalid PDF format
- Corrupted PDF files
- Network errors (when downloading from URLs)
- Page range errors

Always check for errors when using the SDK:

```go
text, err := pdfreader.ExtractTextFromFile("document.pdf")
if err != nil {
    // Handle the error appropriately
    log.Printf("Failed to extract text: %v", err)
    return
}
```

## Dependencies

This SDK uses the following external library:

- `github.com/ledongthuc/pdf` - For PDF parsing and text extraction

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

If you encounter any issues or have questions, please open an issue on the GitHub repository.
