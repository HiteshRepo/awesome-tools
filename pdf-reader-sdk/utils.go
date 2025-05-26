package pdfreader

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func ExtractTextFromFile(filePath string) (string, error) {
	reader, err := NewPDFReader(filePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	return reader.ExtractText(nil)
}

func ExtractTextFromFileWithOptions(filePath string, options *TextExtractOptions) (string, error) {
	reader, err := NewPDFReader(filePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	return reader.ExtractText(options)
}

func ExtractTextFromBytes(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	pdfReader, err := NewPDFReaderFromReader(reader, int64(len(data)))
	if err != nil {
		return "", err
	}
	defer pdfReader.Close()

	return pdfReader.ExtractText(nil)
}

func ExtractTextFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download PDF from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download PDF: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read PDF data: %w", err)
	}

	return ExtractTextFromBytes(data)
}

func ValidatePDFFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".pdf" {
		return fmt.Errorf("file does not have .pdf extension")
	}

	header := make([]byte, 4)
	_, err = file.Read(header)
	if err != nil {
		return fmt.Errorf("failed to read file header: %w", err)
	}

	if string(header) != "%PDF" {
		return fmt.Errorf("file does not appear to be a valid PDF")
	}

	return nil
}

func GetFileInfo(filePath string) (*FileInfo, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	reader, err := NewPDFReader(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	docInfo := reader.GetDocumentInfo()

	return &FileInfo{
		FileName:     filepath.Base(filePath),
		FilePath:     filePath,
		FileSize:     stat.Size(),
		ModTime:      stat.ModTime(),
		PageCount:    docInfo.PageCount,
		DocumentInfo: docInfo,
	}, nil
}

type FileInfo struct {
	FileName     string
	FilePath     string
	FileSize     int64
	ModTime      interface{}
	PageCount    int
	DocumentInfo DocumentInfo
}

func BatchExtractText(filePaths []string) (map[string]string, error) {
	results := make(map[string]string)

	for _, filePath := range filePaths {
		text, err := ExtractTextFromFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to extract text from %s: %w", filePath, err)
		}
		results[filePath] = text
	}

	return results, nil
}

func SearchTextInPDF(filePath, searchText string) ([]PageMatch, error) {
	reader, err := NewPDFReader(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var matches []PageMatch
	pageCount := reader.GetPageCount()

	for pageNum := 1; pageNum <= pageCount; pageNum++ {
		pageText, err := reader.ExtractTextFromPage(pageNum)
		if err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(pageText), strings.ToLower(searchText)) {
			matches = append(matches, PageMatch{
				PageNumber: pageNum,
				Text:       pageText,
			})
		}
	}

	return matches, nil
}

type PageMatch struct {
	PageNumber int
	Text       string
}
