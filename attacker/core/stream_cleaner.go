package core

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
)

// CleanObjectStreamInContent cleans a specific object's stream in the content buffer
func CleanObjectStreamInContent(content []byte, objID int) ([]byte, error) {
	// Convert content to string properly for PDF binary data
	contentStr := string(content)

	// Create pattern to find specific object's stream with DOTALL flag
	pattern := fmt.Sprintf(`%d\s+0\s+obj[\s\S]*?>>\s*stream\s*([\s\S]*?)\s*endstream`, objID)
	objPattern := regexp.MustCompile(pattern)
	matches := objPattern.FindStringSubmatchIndex(contentStr)

	if len(matches) < 4 {
		return content, fmt.Errorf("object %d stream not found", objID)
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
			// Limit output for log readability
			preview := string(decompressed)
			if len(preview) > 50 {
				preview = preview[:50] + "..."
			}
			fmt.Printf("[!] Original decompressed content: %q\n", preview)
		}
	}

		// Create a byte slice of null bytes with the exact original stream length

		emptyCompressed := bytes.Repeat([]byte{0x00}, len(streamContent))

	

	fmt.Printf("[+] New stream length: %d bytes\n", len(emptyCompressed))

	// Build new content
	var newContent []byte
	newContent = append(newContent, content[:streamStart]...)
	newContent = append(newContent, emptyCompressed...)
	newContent = append(newContent, content[streamEnd:]...)

	return newContent, nil
}

// FindSMaskObjects finds all objects referenced as SMask
func FindSMaskObjects(content []byte) []int {
	var smaskIDs []int
	contentStr := string(content)

	// Pattern to find /SMask reference: /SMask 123 0 R
	re := regexp.MustCompile(`/SMask\s+(\d+)\s+0\s+R`)
	matches := re.FindAllStringSubmatch(contentStr, -1)

	for _, match := range matches {
		if len(match) > 1 {
			id, err := strconv.Atoi(match[1])
			if err == nil {
				// Check if ID is already in list to avoid duplicates
				exists := false
				for _, existingID := range smaskIDs {
					if existingID == id {
						exists = true
						break
					}
				}
				if !exists {
					smaskIDs = append(smaskIDs, id)
				}
			}
		}
	}

	return smaskIDs
}

// CleanObjectStream precisely cleans a specific object's stream while preserving PDF structure
// (Legacy wrapper for backward compatibility)
func CleanObjectStream(filePath string, objID int) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	newContent, err := CleanObjectStreamInContent(content, objID)
	if err != nil {
		return "", err
	}

	// Write output
	outputPath := filePath + "_stream_cleaned"
	err = os.WriteFile(outputPath, newContent, 0644)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// StreamCleaner precisely cleans embedded file streams and SMask streams
func StreamCleaner(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// 1. Clean standard Anchor 1 (Object 72)
	fmt.Println("[*] Cleaning Anchor 1 (Object 72)...")
	content, err = CleanObjectStreamInContent(content, 72)
	if err != nil {
		fmt.Printf("[!] Warning: Anchor 1 (Obj 72) issue: %v\n", err)
		// Don't exit, try to continue cleaning other parts
	}

	// 2. Find and Clean SMask Anchors (Anchor 2)
	smaskIDs := FindSMaskObjects(content)
	fmt.Printf("[*] Found %d SMask object(s): %v\n", len(smaskIDs), smaskIDs)

	for _, id := range smaskIDs {
		fmt.Printf("[*] Cleaning potential SMask Anchor (Object %d)...\n", id)
		newContent, err := CleanObjectStreamInContent(content, id)
		if err != nil {
			fmt.Printf("[!] Warning: Failed to clean SMask object %d: %v\n", id, err)
		} else {
			content = newContent
		}
	}

	// Write output
	outputPath := filePath + "_stream_cleaned"
	err = os.WriteFile(outputPath, content, 0644)
	if err != nil {
		return "", err
	}

	return outputPath, nil
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