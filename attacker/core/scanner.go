package core

import (
	"bytes"
	"fmt"
	"os"
)

// ScanResult holds the result of a PDF scan
type ScanResult struct {
	IsSuspicious    bool
	SuspiciousBytes int64
	EOFOffset       int64
	FileSize        int64
}

// FindLastEOF searches for the last occurrence of %%EOF in the file.
// It returns the offset of the beginning of %%EOF.
func FindLastEOF(content []byte) int64 {
	// %%EOF is the marker.
	// We search from the end.
	marker := []byte("%%EOF")
	idx := bytes.LastIndex(content, marker)
	if idx == -1 {
		return -1
	}
	return int64(idx)
}

// ScanPDF checks if a PDF file has suspicious data after the last %%EOF.
func ScanPDF(filePath string) (*ScanResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	fileSize := int64(len(content))
	eofOffset := FindLastEOF(content)

	if eofOffset == -1 {
		return nil, fmt.Errorf("invalid PDF: %%EOF marker not found")
	}

	// Calculate end of EOF marker
	// %%EOF is 5 bytes.
	endOfMarker := eofOffset + 5

	// Calculate trailing bytes
	trailingBytes := fileSize - endOfMarker

	isSuspicious := trailingBytes > 0

	// Refinement: Check if trailing bytes are just whitespace
	// Common PDF implementations might add a newline after %%EOF.
	if isSuspicious {
		trailing := content[endOfMarker:]
		isWhitespace := true
		for _, b := range trailing {
			// 0x0A (LF), 0x0D (CR), 0x20 (Space), 0x09 (Tab)
			if b != '\n' && b != '\r' && b != ' ' && b != '\t' {
				isWhitespace = false
				break
			}
		}
		if isWhitespace {
			isSuspicious = false
		}
	}

	return &ScanResult{
		IsSuspicious:    isSuspicious,
		SuspiciousBytes: trailingBytes,
		EOFOffset:       eofOffset,
		FileSize:        fileSize,
	}, nil
}
