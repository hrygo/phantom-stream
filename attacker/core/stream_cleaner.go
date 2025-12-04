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

// StreamCleaner precisely cleans embedded file streams
func StreamCleaner(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Find object 72 (the embedded file stream)
	obj72Pattern := regexp.MustCompile(`72\s+0\s+obj\s*<</([^>]*)>>\s*stream\s*(.*?)\s*endstream`)
	obj72Matches := obj72Pattern.FindStringSubmatch(string(content))

	if len(obj72Matches) < 3 {
		return "", fmt.Errorf("object 72 not found or invalid format")
	}

	// Parse the dictionary part
	dictStr := obj72Matches[1]
	streamContent := obj72Matches[2]

	// Check if it's compressed
	if strings.Contains(dictStr, "/FlateDecode") {
		// Decompress to see what's inside
		decompressed, err := flateDecode([]byte(streamContent))
		if err == nil {
			fmt.Printf("[!] Original stream content: %s\n", string(decompressed))
		}
	}

	// Replace with empty compressed content
	emptyStream := compressData([]byte{})

	// Build the replacement object
	replacement := fmt.Sprintf(`72 0 obj<</%s>>\nstream\n%s\nendstream`, dictStr, emptyStream)

	// Replace the entire object 72
	modifiedContent := bytes.ReplaceAll(content, []byte(obj72Matches[0]), []byte(replacement))

	// Write the modified file
	outputPath := filePath + "_stream_cleaned"
	err = os.WriteFile(outputPath, modifiedContent, 0644)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// flateDecode decompresses flate-encoded data
func flateDecode(data []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(data))
	defer r.Close()

	result, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// compressData compresses data with flate
func compressData(data []byte) []byte {
	var buf bytes.Buffer
	w, _ := flate.NewWriter(&buf, flate.DefaultCompression)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}

// CleanEmbeddedFilesReference removes only the EmbeddedFiles entry from Catalog
func CleanEmbeddedFilesReference(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Find the Catalog object - try different patterns
	catalogPatterns := []string{
		`(\d+\s+0\s+obj\s*<</Type/Catalog.*?)>>`,
		`(\d+\s+0\s+obj\s*<</.*?/Type/Catalog.*?)>>`,
		`1\s+0\s+obj\s*(<</.*?/Type/Catalog.*?)>>`,
	}

	var catalogContent string

	for _, pattern := range catalogPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(string(content))
		if len(matches) >= 2 {
			catalogContent = matches[1]
			break
		}
	}

	if catalogContent == "" {
		return "", fmt.Errorf("catalog not found")
	}

	// Remove EmbeddedFiles from catalog
	if strings.Contains(catalogContent, "/EmbeddedFiles") {
		// Remove the entire EmbeddedFiles entry
		embeddedFilesPattern := regexp.MustCompile(`/EmbeddedFiles\s*<<.*?>>>`)
		catalogContent = embeddedFilesPattern.ReplaceAllString(catalogContent, "")

		// Remove trailing whitespace
		catalogContent = strings.TrimSpace(catalogContent)

		// Rebuild the catalog object
		newCatalog := catalogContent + ">>"
		// Find the original catalog object to replace
		originalCatalog := regexp.MustCompile(regexp.QuoteMeta(catalogContent) + ">>")
		replacement := originalCatalog.ReplaceAllString(string(content), newCatalog)

		// Write the modified file
		outputPath := filePath + "_catalog_cleaned"
		err = os.WriteFile(outputPath, []byte(replacement), 0644)
		if err != nil {
			return "", err
		}

		return outputPath, nil
	}

	return "", fmt.Errorf("no EmbeddedFiles found in catalog")
}