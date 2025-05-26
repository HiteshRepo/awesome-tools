package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	pdfreader "github.com/HiteshRepo/awesome-tools/pdf-reader"
)

func main() {
	fileLocFlag := flag.String("file-loc", "", "location of PDF")
	flag.Parse()

	fileLoc := ""
	if fileLocFlag != nil {
		fileLoc = *fileLocFlag
	}

	if len(fileLoc) == 0 {
		flag.Usage()
		return
	}

	if _, err := os.Stat(fileLoc); os.IsNotExist(err) {
		log.Fatalf("file does not exist: %s", fileLoc)
	}

	fmt.Println("=== Example 1: Simple Text Extraction ===")
	text, err := pdfreader.ExtractTextFromFile(fileLoc)
	if err != nil {
		log.Fatalf("Failed to extract text: %v", err)
	}
	fmt.Printf("Extracted text (first 200 chars): %s...\n\n", truncateString(text, 200))

	fmt.Println("=== Example 2: Document Information ===")
	fileInfo, err := pdfreader.GetFileInfo(fileLoc)
	if err != nil {
		log.Fatalf("Failed to get file info: %v", err)
	}

	fmt.Printf("File: %s\n", fileInfo.FileName)
	fmt.Printf("Size: %d bytes\n", fileInfo.FileSize)
	fmt.Printf("Pages: %d\n", fileInfo.PageCount)
	fmt.Printf("Title: %s\n", fileInfo.DocumentInfo.Title)
	fmt.Printf("Author: %s\n", fileInfo.DocumentInfo.Author)
	fmt.Printf("Creator: %s\n", fileInfo.DocumentInfo.Creator)
	fmt.Printf("Producer: %s\n", fileInfo.DocumentInfo.Producer)
	fmt.Println()

	fmt.Println("=== Example 3: Extract Text from Specific Pages ===")
	reader, err := pdfreader.NewPDFReader(fileLoc)
	if err != nil {
		log.Fatalf("Failed to create PDF reader: %v", err)
	}
	defer reader.Close()

	pageCount := reader.GetPageCount()
	if pageCount > 0 {
		pageText, err := reader.ExtractTextFromPage(1)
		if err != nil {
			log.Printf("Failed to extract text from page 1: %v", err)
		} else {
			fmt.Printf("Page 1 text (first 150 chars): %s...\n", truncateString(pageText, 150))
		}
	}
	fmt.Println()

	fmt.Println("=== Example 4: Extract Text with Custom Options ===")
	options := &pdfreader.TextExtractOptions{
		PageRange: &pdfreader.PageRange{
			Start: 1,
			End:   2,
		},
		JoinLines:          true,
		PreserveFormatting: false,
	}

	customText, err := pdfreader.ExtractTextFromFileWithOptions(fileLoc, options)
	if err != nil {
		log.Printf("Failed to extract text with options: %v", err)
	} else {
		fmt.Printf("Custom extraction (first 200 chars): %s...\n", truncateString(customText, 200))
	}
	fmt.Println()

	fmt.Println("=== Example 5: Search Text in PDF ===")
	searchTerm := "the"
	matches, err := pdfreader.SearchTextInPDF(fileLoc, searchTerm)
	if err != nil {
		log.Printf("Failed to search text: %v", err)
	} else {
		fmt.Printf("Found '%s' on %d pages:\n", searchTerm, len(matches))
		for i, match := range matches {
			if i >= 3 {
				fmt.Printf("... and %d more pages\n", len(matches)-3)
				break
			}
			fmt.Printf("  Page %d\n", match.PageNumber)
		}
	}
	fmt.Println()

	fmt.Println("=== Example 6: Validate PDF File ===")
	err = pdfreader.ValidatePDFFile(fileLoc)
	if err != nil {
		fmt.Printf("PDF validation failed: %v\n", err)
	} else {
		fmt.Println("PDF file is valid")
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
