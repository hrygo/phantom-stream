package core

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"os"
	"regexp"
)

// CleanObjectStream precisely cleans a specific object's stream while preserving PDF structure
func CleanObjectStream(filePath string, objID int) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Convert content to string properly for PDF binary data
	contentStr := string(content)

	// Create pattern to find specific object's stream with DOTALL flag
	pattern := fmt.Sprintf(`%d\s+0\s+obj[\s\S]*?>>\s*stream\s*([\s\S]*?)\s*endstream`, objID)
	objPattern := regexp.MustCompile(pattern)
	matches := objPattern.FindStringSubmatchIndex(contentStr)

	if len(matches) < 4 {
		return "", fmt.Errorf("object %d stream not found", objID)
	}

	fmt.Printf("[+] Found object %d stream\n", objID)

	// Extract the stream content (only one capture group)
	streamStart := matches[2]
	streamEnd := matches[3]
	streamContent := content[streamStart:streamEnd]

	fmt.Printf("[+] Original stream length: %d bytes\n", len(streamContent))

	// Try to decompress to see what we're removing
	if len(streamContent) > 0 {
		reader := flate.NewReader(bytes.NewReader(streamContent))
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err == nil {
			fmt.Printf("[!] Original decompressed content: %q\n", string(decompressed))
		}
	}

	// Create empty compressed data that matches exact size
	var buf bytes.Buffer
	w, _ := flate.NewWriter(&buf, flate.NoCompression)
	w.Write([]byte{})  // Write empty content
	w.Close()

	emptyCompressed := buf.Bytes()

	// Pad or truncate to match original size
	if len(emptyCompressed) < len(streamContent) {
		padded := make([]byte, len(streamContent))
		copy(padded, emptyCompressed)
		emptyCompressed = padded
	} else if len(emptyCompressed) > len(streamContent) {
		emptyCompressed = emptyCompressed[:len(streamContent)]
	}

	fmt.Printf("[+] New stream length: %d bytes\n", len(emptyCompressed))

	// Build new content
	var newContent []byte
	newContent = append(newContent, content[:streamStart]...)
	newContent = append(newContent, emptyCompressed...)
	newContent = append(newContent, content[streamEnd:]...)

	// Write output
	outputPath := filePath + "_stream_cleaned"
	err = os.WriteFile(outputPath, newContent, 0644)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// StreamCleaner precisely cleans embedded file streams (defaults to object 72)
func StreamCleaner(filePath string) (string, error) {
	return CleanObjectStream(filePath, 72)
}

// VerifyPDFIntegrity checks if a PDF file is structurally valid
func VerifyPDFIntegrity(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file: %v", err)
	}

	// Check PDF header
	if !bytes.HasPrefix(content, []byte("%PDF")) {
		return fmt.Errorf("invalid PDF header")
	}

	// Check EOF marker
	if !bytes.HasSuffix(bytes.TrimSpace(content), []byte("%%EOF")) {
		return fmt.Errorf("missing EOF marker")
	}

	// Count xref tables
	xrefCount := bytes.Count(content, []byte("xref"))
	if xrefCount == 0 {
		return fmt.Errorf("no xref table found")
	}

	fmt.Printf("[+] PDF integrity verification passed\n")
	fmt.Printf("[+] Found %d xref table(s)\n", xrefCount)

	return nil
}