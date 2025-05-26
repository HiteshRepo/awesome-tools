package pdfreader

import (
	"os"
	"strings"
	"testing"
)

const testPDFPath = "test-data/Hitesh-Pattanayak-Resume-2024.pdf"

func TestNewPDFReader(t *testing.T) {
	reader, err := NewPDFReader(testPDFPath)
	if err != nil {
		t.Fatalf("Failed to create PDF reader: %v", err)
	}

	if reader == nil {
		t.Fatal("PDF reader is nil")
	}

	pageCount := reader.GetPageCount()
	if pageCount <= 0 {
		t.Errorf("Expected positive page count, got %d", pageCount)
	}

	reader.Close()
}

func TestNewPDFReaderFromReader(t *testing.T) {
	file, err := os.Open(testPDFPath)
	if err != nil {
		t.Fatalf("Failed to open test PDF: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to get file stats: %v", err)
	}

	pdfReader, err := NewPDFReaderFromReader(file, stat.Size())
	if err != nil {
		t.Fatalf("Failed to create PDF reader: %v", err)
	}

	if pdfReader == nil {
		t.Fatal("PDF reader is nil")
	}

	pageCount := pdfReader.GetPageCount()
	if pageCount <= 0 {
		t.Errorf("Expected positive page count, got %d", pageCount)
	}
}

func TestExtractText(t *testing.T) {
	reader, err := NewPDFReader(testPDFPath)
	if err != nil {
		t.Fatalf("Failed to create PDF reader: %v", err)
	}
	defer reader.Close()

	text, err := reader.ExtractText(nil)
	if err != nil {
		t.Fatalf("Failed to extract text: %v", err)
	}

	if len(text) == 0 {
		t.Error("Expected non-empty text extraction")
	}
}

func TestExtractTextWithOptions(t *testing.T) {
	reader, err := NewPDFReader(testPDFPath)
	if err != nil {
		t.Fatalf("Failed to create PDF reader: %v", err)
	}
	defer reader.Close()

	options := &TextExtractOptions{
		PageRange: &PageRange{Start: 1, End: 1},
		JoinLines: true,
	}

	text, err := reader.ExtractText(options)
	if err != nil {
		t.Fatalf("Failed to extract text with options: %v", err)
	}

	if len(text) == 0 {
		t.Error("Expected non-empty text extraction")
	}
}

func TestExtractTextFromPage(t *testing.T) {
	reader, err := NewPDFReader(testPDFPath)
	if err != nil {
		t.Fatalf("Failed to create PDF reader: %v", err)
	}
	defer reader.Close()

	pageCount := reader.GetPageCount()
	if pageCount <= 0 {
		t.Skip("No pages to test")
	}

	pageText, err := reader.ExtractTextFromPage(1)
	if err != nil {
		t.Fatalf("Failed to extract text from page 1: %v", err)
	}

	if len(pageText) == 0 {
		t.Error("Expected non-empty page text")
	}
}

func TestExtractTextFromPageOutOfRange(t *testing.T) {
	reader, err := NewPDFReader(testPDFPath)
	if err != nil {
		t.Fatalf("Failed to create PDF reader: %v", err)
	}
	defer reader.Close()

	pageCount := reader.GetPageCount()
	_, err = reader.ExtractTextFromPage(pageCount + 1)
	if err == nil {
		t.Error("Expected error for out of range page number")
	}
}

func TestGetDocumentInfo(t *testing.T) {
	reader, err := NewPDFReader(testPDFPath)
	if err != nil {
		t.Fatalf("Failed to create PDF reader: %v", err)
	}
	defer reader.Close()

	info := reader.GetDocumentInfo()
	if info.PageCount <= 0 {
		t.Errorf("Expected positive page count, got %d", info.PageCount)
	}
}

func TestExtractTextFromFile(t *testing.T) {
	text, err := ExtractTextFromFile(testPDFPath)
	if err != nil {
		t.Fatalf("Failed to extract text from file: %v", err)
	}

	if len(text) == 0 {
		t.Error("Expected non-empty text extraction")
	}
}

func TestExtractTextFromFileWithOptions(t *testing.T) {
	options := &TextExtractOptions{
		PageRange: &PageRange{Start: 1, End: 1},
		JoinLines: true,
	}

	text, err := ExtractTextFromFileWithOptions(testPDFPath, options)
	if err != nil {
		t.Fatalf("Failed to extract text from file with options: %v", err)
	}

	if len(text) == 0 {
		t.Error("Expected non-empty text extraction")
	}
}

func TestGetFileInfo(t *testing.T) {
	fileInfo, err := GetFileInfo(testPDFPath)
	if err != nil {
		t.Fatalf("Failed to get file info: %v", err)
	}

	if fileInfo.PageCount <= 0 {
		t.Errorf("Expected positive page count, got %d", fileInfo.PageCount)
	}

	if fileInfo.FileSize <= 0 {
		t.Errorf("Expected positive file size, got %d", fileInfo.FileSize)
	}

	if fileInfo.FileName == "" {
		t.Error("Expected non-empty file name")
	}
}

func TestSearchTextInPDF(t *testing.T) {
	matches, err := SearchTextInPDF(testPDFPath, "Hitesh")
	if err != nil {
		t.Fatalf("Failed to search text in PDF: %v", err)
	}

	if len(matches) == 0 {
		t.Error("Expected to find 'Hitesh' in the resume PDF")
	}

	for _, match := range matches {
		if match.PageNumber <= 0 {
			t.Errorf("Expected positive page number, got %d", match.PageNumber)
		}
		if len(match.Text) == 0 {
			t.Error("Expected non-empty match text")
		}
	}
}

func TestValidatePDFFile_ValidPDF(t *testing.T) {
	err := ValidatePDFFile(testPDFPath)
	if err != nil {
		t.Errorf("Expected no error for valid PDF, got: %v", err)
	}
}

func TestValidatePDFFile_InvalidExtension(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	err = ValidatePDFFile(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for non-PDF file extension")
	}

	if !strings.Contains(err.Error(), "does not have .pdf extension") {
		t.Errorf("Expected extension error, got: %v", err)
	}
}

func TestValidatePDFFile_InvalidHeader(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test*.pdf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("INVALID")
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	err = ValidatePDFFile(tmpFile.Name())
	if err == nil {
		t.Error("Expected error for invalid PDF header")
	}

	if !strings.Contains(err.Error(), "does not appear to be a valid PDF") {
		t.Errorf("Expected header validation error, got: %v", err)
	}
}

func TestBatchExtractText(t *testing.T) {
	files := []string{testPDFPath}
	results, err := BatchExtractText(files)
	if err != nil {
		t.Fatalf("Failed to batch extract text: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	text, exists := results[testPDFPath]
	if !exists {
		t.Error("Expected result for test PDF file")
	}

	if len(text) == 0 {
		t.Error("Expected non-empty text extraction")
	}
}

func TestBatchExtractText_EmptySlice(t *testing.T) {
	results, err := BatchExtractText([]string{})
	if err != nil {
		t.Errorf("Expected no error for empty slice, got: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected empty results for empty input, got %d results", len(results))
	}
}

func TestPageRange(t *testing.T) {
	tests := []struct {
		name      string
		pageRange *PageRange
		pageCount int
		wantStart int
		wantEnd   int
	}{
		{
			name:      "nil page range",
			pageRange: nil,
			pageCount: 10,
			wantStart: 1,
			wantEnd:   10,
		},
		{
			name:      "valid range",
			pageRange: &PageRange{Start: 2, End: 5},
			pageCount: 10,
			wantStart: 2,
			wantEnd:   5,
		},
		{
			name:      "end is -1",
			pageRange: &PageRange{Start: 1, End: -1},
			pageCount: 10,
			wantStart: 1,
			wantEnd:   10,
		},
		{
			name:      "start less than 1",
			pageRange: &PageRange{Start: 0, End: 5},
			pageCount: 10,
			wantStart: 1,
			wantEnd:   5,
		},
		{
			name:      "end greater than page count",
			pageRange: &PageRange{Start: 1, End: 15},
			pageCount: 10,
			wantStart: 1,
			wantEnd:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startPage := 1
			endPage := tt.pageCount

			if tt.pageRange != nil {
				startPage = tt.pageRange.Start
				if startPage < 1 {
					startPage = 1
				}

				if tt.pageRange.End == -1 || tt.pageRange.End > tt.pageCount {
					endPage = tt.pageCount
				} else {
					endPage = tt.pageRange.End
				}
			}

			if startPage != tt.wantStart {
				t.Errorf("Expected start page %d, got %d", tt.wantStart, startPage)
			}

			if endPage != tt.wantEnd {
				t.Errorf("Expected end page %d, got %d", tt.wantEnd, endPage)
			}
		})
	}
}

func BenchmarkExtractTextFromFile(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractTextFromFile(testPDFPath)
	}
}

func BenchmarkValidatePDFFile(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidatePDFFile(testPDFPath)
	}
}
