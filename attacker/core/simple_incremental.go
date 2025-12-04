package core

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// SimpleIncrementalCleaner provides a simple approach to modify PDF streams
func SimpleIncrementalClean(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	contentStr := string(content)

	// Find object 72 - more flexible pattern
	obj72Pattern := regexp.MustCompile(`72\s+0\s+obj\s*<</[^>]*>>\s*stream\s*(.*?)\s*endstream`)
	matches := obj72Pattern.FindStringSubmatchIndex(contentStr)

	if len(matches) < 4 {
		return "", fmt.Errorf("object 72 stream not found")
	}

	before := contentStr[:matches[0]]
	dict := contentStr[matches[2]:matches[3]]
	streamContent := contentStr[matches[4]:matches[5]]
	matchEnd := matches[0] + (matches[5]-matches[0]) + len("endstream")
	after := contentStr[matchEnd:]

	fmt.Printf("[+] Found object 72 stream\n")
	fmt.Printf("[+] Dictionary: %s\n", dict)

	// Try to decompress original
	reader := flate.NewReader(strings.NewReader(streamContent))
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err == nil {
		fmt.Printf("[!] Original content: %q\n", string(decompressed))
	}

	// Create minimal compressed data (empty content)
	var buf bytes.Buffer
	w, _ := flate.NewWriter(&buf, flate.NoCompression)
	w.Write([]byte{})
	w.Close()
	minimalCompressed := buf.Bytes()

	// If minimal compressed is smaller, pad with zeros to match original size
	targetLen := len(streamContent)
	if len(minimalCompressed) < targetLen {
		padded := make([]byte, targetLen)
		copy(padded, minimalCompressed)
		minimalCompressed = padded
	} else if len(minimalCompressed) > targetLen {
		minimalCompressed = minimalCompressed[:targetLen]
	}

	// Build the updated content
	var updatedContent []byte
	updatedContent = append(updatedContent, []byte(before)...)
	updatedContent = append(updatedContent, []byte(dict)...)
	updatedContent = append(updatedContent, []byte("\nstream\n")...)
	updatedContent = append(updatedContent, minimalCompressed...)
	updatedContent = append(updatedContent, []byte("\nendstream\nendobj\n")...)
	updatedContent = append(updatedContent, []byte(after)...)

	// Write the updated file
	outputPath := filePath + "_incremental_cleaned"
	err = os.WriteFile(outputPath, updatedContent, 0644)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// TestIncrementalClean tests the incremental clean approach
func TestIncrementalClean(filePath string) error {
	fmt.Printf("\n=== Testing Incremental Clean on %s ===\n", filePath)

	// First, clean the file
	outputPath, err := SimpleIncrementalClean(filePath)
	if err != nil {
		return fmt.Errorf("incremental clean failed: %v", err)
	}

	// Read original content for comparison
	originalContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read original file: %v", err)
	}

	// Verify the output
	outputContent, err := os.ReadFile(outputPath)
	if err != nil {
		return fmt.Errorf("cannot read output file: %v", err)
	}

	// Check if it's still a valid PDF
	if !bytes.HasPrefix(outputContent, []byte("%PDF")) {
		return fmt.Errorf("invalid PDF header in output")
	}

	if !bytes.HasSuffix(bytes.TrimSpace(outputContent), []byte("%%EOF")) {
		return fmt.Errorf("missing EOF marker in output")
	}

	fmt.Printf("[+] PDF header valid\n")
	fmt.Printf("[+] EOF marker found\n")
	fmt.Printf("[+] Original size: %d bytes\n", len(originalContent))
	fmt.Printf("[+] Cleaned size: %d bytes\n", len(outputContent))
	fmt.Printf("[+] Size difference: %d bytes\n", len(outputContent)-len(originalContent))

	fmt.Printf("[+] Incremental clean completed!\n")
	fmt.Printf("[+] Output file: %s\n", outputPath)

	return nil
}