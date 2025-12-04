package core

import (
	"bytes"
	"compress/flate"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// SafeCleaner performs safe cleaning without corrupting the PDF
type SafeCleaner struct {
	FilePath string
	Content  []byte
}

// NewSafeCleaner creates a new safe cleaner instance
func NewSafeCleaner(filePath string) (*SafeCleaner, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return &SafeCleaner{
		FilePath: filePath,
		Content:  content,
	}, nil
}

// CleanStreamContent safely cleans stream content
func (sc *SafeCleaner) CleanStreamContent() (string, error) {
	// Find object 72 with stream content
	obj72Pattern := regexp.MustCompile(`(72\s+0\s+obj\s*<</[^>]*>>\s*stream\s*)(.*?)(\s*endstream)`)
	matches := obj72Pattern.FindStringSubmatchIndex(string(sc.Content))

	if len(matches) < 6 {
		return "", fmt.Errorf("object 72 stream not found")
	}

	// Extract the three parts
	before := matches[0]             // Start of object
	streamStart := matches[2]        // Start of stream content
	streamEnd := matches[4]          // End of stream content
	after := matches[5]              // Start after endstream

	// Get the actual stream content
	contentBytes := sc.Content[streamStart:streamEnd]

	// Try to decompress to see what it is
	decompressed, err := flateDecode(contentBytes)
	if err == nil {
		fmt.Printf("[!] Original stream content: %q\n", string(decompressed))
	}

	// Create minimal valid flate data
	emptyData := []byte{}
	var buf bytes.Buffer
	w, _ := flate.NewWriter(&buf, flate.DefaultCompression)
	w.Write(emptyData)
	w.Close()
	replacement := buf.Bytes()

	// Build the new content
	var newContent []byte
	newContent = append(newContent, sc.Content[:before]...)
	newContent = append(newContent, []byte("\n")...)          // Ensure newline after 'stream'
	newContent = append(newContent, replacement...)           // New compressed data
	newContent = append(newContent, sc.Content[after:]...)   // Rest of file

	// Write to output file
	outputPath := sc.FilePath + "_safe_processed"
	if strings.HasSuffix(sc.FilePath, ".pdf") {
		baseName := sc.FilePath[:len(sc.FilePath)-4]
		outputPath = baseName + "_safe_processed.pdf"
	}

	err = os.WriteFile(outputPath, newContent, 0644)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// Alternative clean method - remove EmbeddedFiles reference
func (sc *SafeCleaner) RemoveEmbeddedFilesRef() (string, error) {
	contentStr := string(sc.Content)

	// Find the EmbeddedFiles reference in the catalog
	embeddedFilesPattern := regexp.MustCompile(`/EmbeddedFiles\s*<</Names\s*\[<666f6e745f6c6963656e73652e747874>\s*73\s+0\s+R\s*\]>>>>`)

	if !embeddedFilesPattern.MatchString(contentStr) {
		return "", fmt.Errorf("EmbeddedFiles reference not found")
	}

	// Remove only the EmbeddedFiles entry, keep everything else
	cleanedContent := embeddedFilesPattern.ReplaceAllString(contentStr, "")

	// Write to output file
	outputPath := sc.FilePath + "_cleaned_ref"
	if strings.HasSuffix(sc.FilePath, ".pdf") {
		baseName := sc.FilePath[:len(sc.FilePath)-4]
		outputPath = baseName + "_cleaned_ref.pdf"
	}

	err := os.WriteFile(outputPath, []byte(cleanedContent), 0644)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

