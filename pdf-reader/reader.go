package pdfreader

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

type PDFReader struct {
	filePath string
	file     *pdf.Reader
	osFile   *os.File
}

type TextExtractOptions struct {
	PageRange          *PageRange
	PreserveFormatting bool
	JoinLines          bool
}

type PageRange struct {
	Start int
	End   int
}

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

func NewPDFReader(filePath string) (*PDFReader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open PDF file: %w", err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	pdfReader, err := pdf.NewReader(file, fileInfo.Size())
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to create PDF reader: %w", err)
	}

	return &PDFReader{
		filePath: filePath,
		file:     pdfReader,
		osFile:   file,
	}, nil
}

func NewPDFReaderFromReader(reader io.ReaderAt, size int64) (*PDFReader, error) {
	pdfReader, err := pdf.NewReader(reader, size)
	if err != nil {
		return nil, fmt.Errorf("failed to create PDF reader: %w", err)
	}

	return &PDFReader{
		file: pdfReader,
	}, nil
}

func (pr *PDFReader) GetDocumentInfo() DocumentInfo {
	info := DocumentInfo{
		PageCount: pr.file.NumPage(),
	}

	trailer := pr.file.Trailer()
	if !trailer.IsNull() {
		infoDict := trailer.Key("Info")
		if !infoDict.IsNull() {
			if title := infoDict.Key("Title"); !title.IsNull() {
				info.Title = title.String()
			}
			if author := infoDict.Key("Author"); !author.IsNull() {
				info.Author = author.String()
			}
			if subject := infoDict.Key("Subject"); !subject.IsNull() {
				info.Subject = subject.String()
			}
			if creator := infoDict.Key("Creator"); !creator.IsNull() {
				info.Creator = creator.String()
			}
			if producer := infoDict.Key("Producer"); !producer.IsNull() {
				info.Producer = producer.String()
			}
			if creationDate := infoDict.Key("CreationDate"); !creationDate.IsNull() {
				info.CreationDate = creationDate.String()
			}
			if modDate := infoDict.Key("ModDate"); !modDate.IsNull() {
				info.ModDate = modDate.String()
			}
		}
	}

	return info
}

func (pr *PDFReader) ExtractText(options *TextExtractOptions) (string, error) {
	if options == nil {
		options = &TextExtractOptions{
			JoinLines: true,
		}
	}

	var result strings.Builder
	pageCount := pr.file.NumPage()

	startPage := 1
	endPage := pageCount

	if options.PageRange != nil {
		startPage = options.PageRange.Start
		if startPage < 1 {
			startPage = 1
		}

		if options.PageRange.End == -1 || options.PageRange.End > pageCount {
			endPage = pageCount
		} else {
			endPage = options.PageRange.End
		}
	}

	for pageNum := startPage; pageNum <= endPage; pageNum++ {
		page := pr.file.Page(pageNum)
		if page.V.IsNull() {
			continue
		}

		pageText, err := page.GetPlainText(nil)
		if err != nil {
			return "", fmt.Errorf("failed to extract text from page %d: %w", pageNum, err)
		}

		if options.JoinLines {
			pageText = strings.Join(strings.Fields(pageText), " ")
		}

		if pageText != "" {
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			result.WriteString(pageText)
		}
	}

	return result.String(), nil
}

func (pr *PDFReader) ExtractTextFromPage(pageNum int) (string, error) {
	if pageNum < 1 || pageNum > pr.file.NumPage() {
		return "", fmt.Errorf("page number %d is out of range (1-%d)", pageNum, pr.file.NumPage())
	}

	page := pr.file.Page(pageNum)
	if page.V.IsNull() {
		return "", nil
	}

	pageText, err := page.GetPlainText(nil)
	if err != nil {
		return "", fmt.Errorf("failed to extract text from page %d: %w", pageNum, err)
	}

	return pageText, nil
}

func (pr *PDFReader) GetPageCount() int {
	return pr.file.NumPage()
}

func (pr *PDFReader) Close() error {
	if pr.osFile != nil {
		return pr.osFile.Close()
	}
	return nil
}
