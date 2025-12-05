package core

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
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
		// Note: If the stream is valid zlib, flate.NewReader usually works for raw deflate.
		// For standard zlib wrapped stream, zlib.NewReader is better, but we keep flate for now
		// as we are just inspecting.
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

	// Create a valid empty zlib stream to satisfy "Legal PDF Structure"
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write([]byte{}) // Write empty content
	w.Close()
	validZlib := b.Bytes()

	// Pad with null bytes to match exact original size
	emptyCompressed := make([]byte, len(streamContent))
	if len(validZlib) <= len(streamContent) {
		copy(emptyCompressed, validZlib)
		// The rest is already 0x00 (padding)
	} else {
		// This case should be rare for empty data vs typical stream size
		// If valid zlib > original size, we must truncate (invalid) or panic.
		// But empty zlib is ~8 bytes. Original streams > 60 bytes. Safe.
		fmt.Printf("[!] Warning: Valid zlib header (%d) larger than original stream (%d). Truncating.\n", len(validZlib), len(streamContent))
		copy(emptyCompressed, validZlib[:len(streamContent)])
	}

	fmt.Printf("[+] New stream length: %d bytes (Valid Zlib + Padding)\n", len(emptyCompressed))

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

// SanitizeImageStreamInContent cleans an image/SMask stream by stripping LSBs
func SanitizeImageStreamInContent(content []byte, objID int) ([]byte, error) {
	// Convert content to string properly for PDF binary data
	contentStr := string(content)

	// Create pattern to find specific object's stream
	pattern := fmt.Sprintf(`%d\s+0\s+obj[\s\S]*?>>\s*stream\s*([\s\S]*?)\s*endstream`, objID)
	objPattern := regexp.MustCompile(pattern)
	matches := objPattern.FindStringSubmatchIndex(contentStr)

	if len(matches) < 4 {
		return content, fmt.Errorf("object %d stream not found", objID)
	}

	fmt.Printf("[*] Sanitizing Image/SMask Object %d...\n", objID)

	// Extract stream
	streamStart := matches[2]
	streamEnd := matches[3]
	streamContent := content[streamStart:streamEnd]
	originalLen := len(streamContent)
	fmt.Printf("[+] Original stream length: %d bytes\n", originalLen)

	// SKIP VERY SMALL STREAMS (Structural Masks)
	// Likely 1x1 pixel masks or helper objects. Wiping these causes blank pages.
	if originalLen < 100 {
		fmt.Printf("[!] Stream too small (%d bytes). Skipping to preserve structure.\n", originalLen)
		return content, nil
	}

	// 1. Decompress
	var decompressed []byte
	rc, err := zlib.NewReader(bytes.NewReader(streamContent))
	if err == nil {
		decompressed, err = io.ReadAll(rc)
		rc.Close()
	}

	// If zlib fails, try flate (raw)
	if err != nil || len(decompressed) == 0 {
		fr := flate.NewReader(bytes.NewReader(streamContent))
		decompressed, err = io.ReadAll(fr)
		fr.Close()
	}

	if err != nil {
		fmt.Printf("[!] Warning: Failed to decompress object %d. Skipping LSB sanitization.\n", objID)
		return content, nil // Skip if we can't decompress
	}

	fmt.Printf("[+] Decompressed size: %d bytes\n", len(decompressed))

	// 2. Lossless Canonicalization (Recompression)
	// Instead of LSB quantization (which risks breaking 1-bit masks or indexed colors),
	// we simply recompress the raw pixel data. This strips any "Appended Data"
	// (hidden after the Zlib stream but before endstream) and normalizes compression.
	// This preserves 100% visual fidelity while removing structural hiding places.

	var b bytes.Buffer
	w, _ := zlib.NewWriterLevel(&b, zlib.BestCompression)
	w.Write(decompressed)
	w.Close()
	canonicalCompressed := b.Bytes()

	fmt.Printf("[*] Canonical Recompression: %d -> %d bytes\n", originalLen, len(canonicalCompressed))

	// 3. Validate & Pad
	finalStream := make([]byte, originalLen)

	if len(canonicalCompressed) <= originalLen {
		copy(finalStream, canonicalCompressed)
		// The rest is padding (0x00)
		fmt.Printf("[+] Success: Canonical stream fits.\n")
	} else {
		// If recompression is larger, we CANNOT fit it in the original space without breaking XRef.
		// In this case, we must fallback to keeping the original stream (Risk: Payload stays).
		// However, for most steganography, the payload *adds* entropy/size, so removing it *should* shrink.
		// If it grows, it means the original was highly optimized or we used wrong settings.
		// Safe fallback: Keep original.
		fmt.Printf("[!] Warning: Recompression larger than original. Keeping original stream to preserve structure.\n")
		return content, nil
	}

	// Build new content
	var newContent []byte
	newContent = append(newContent, content[:streamStart]...)
	newContent = append(newContent, finalStream...)
	newContent = append(newContent, content[streamEnd:]...)

	return newContent, nil
}

// StreamCleaner precisely cleans embedded file streams and SMask streams
func StreamCleaner(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// 1. Clean standard Anchor 1 (Object 72) - Using Replacement (Wipe)
	fmt.Println("[*] Cleaning Anchor 1 (Object 72)...")
	content, err = CleanObjectStreamInContent(content, 72)
	if err != nil {
		fmt.Printf("[!] Warning: Anchor 1 (Obj 72) issue: %v\n", err)
	}

	// 2. Clean Anchor 3 (Object 82 - Visual Watermark) - Using Replacement (Wipe)
	fmt.Println("[*] Cleaning Anchor 3 (Object 82 - Visual Watermark)...")
	content, err = CleanObjectStreamInContent(content, 82)
	if err != nil {
		fmt.Printf("[!] Warning: Anchor 3 (Obj 82) issue: %v\n", err)
	}

	// 3. Clean Object 338 (Found 'b78b' watermark) - Using Replacement (Wipe)
	fmt.Println("[*] Cleaning Object 338 (Found 'b78b' watermark)...")
	content, err = CleanObjectStreamInContent(content, 338)
	if err != nil {
		fmt.Printf("[!] Warning: Object 338 issue: %v\n", err)
	}

	// 4. Find and Clean SMask Anchors (Anchor 2) - Using LSB Sanitization
	smaskIDs := FindSMaskObjects(content)
	fmt.Printf("[*] Found %d SMask object(s): %v\n", len(smaskIDs), smaskIDs)

	for _, id := range smaskIDs {
		content, err = SanitizeImageStreamInContent(content, id)
		if err != nil {
			fmt.Printf("[!] Warning: Failed to clean SMask object %d: %v\n", id, err)
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
