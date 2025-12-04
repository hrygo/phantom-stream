package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CleanPDF removes data after the last %%EOF and saves to a new file.
// It returns the path of the cleaned file and the number of bytes removed.
func CleanPDF(filePath string) (string, int64, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read file: %w", err)
	}

	eofOffset := FindLastEOF(content)
	if eofOffset == -1 {
		return "", 0, fmt.Errorf("invalid PDF: %%EOF marker not found")
	}

	// Determine cutoff point.
	// We want to keep %%EOF (5 bytes).
	cutoff := eofOffset + 5

	// Check if there is a newline immediately after %%EOF and preserve it if so.
	// This helps with some readers that expect a trailing newline.
	// We only preserve one sequence of CRLF or LF or CR.
	if int64(len(content)) > cutoff {
		if content[cutoff] == '\r' {
			cutoff++
			if int64(len(content)) > cutoff && content[cutoff] == '\n' {
				cutoff++
			}
		} else if content[cutoff] == '\n' {
			cutoff++
		}
	}

	cleanedContent := content[:cutoff]
	removedBytes := int64(len(content)) - int64(len(cleanedContent))

	// Generate output filename
	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	name := strings.TrimSuffix(filepath.Base(filePath), ext)
	outPath := filepath.Join(dir, fmt.Sprintf("%s_cleaned%s", name, ext))

	// Write to new file
	err = os.WriteFile(outPath, cleanedContent, 0644)
	if err != nil {
		return "", 0, fmt.Errorf("failed to write cleaned file: %w", err)
	}

	return outPath, removedBytes, nil
}
