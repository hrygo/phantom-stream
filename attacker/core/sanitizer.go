package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SanitizeGaps overwrites suspicious data in gaps with spaces.
func SanitizeGaps(filePath string) (string, int64, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read file: %w", err)
	}

	gaps, err := FindSuspiciousGaps(content)
	if err != nil {
		return "", 0, fmt.Errorf("failed to scan for gaps: %w", err)
	}

	if len(gaps) == 0 {
		return "", 0, nil // Nothing to clean
	}

	sanitizedContent := make([]byte, len(content))
	copy(sanitizedContent, content)

	var sanitizedBytes int64

	for _, gap := range gaps {
		// Overwrite gap with spaces (0x20)
		for i := gap.Start; i < gap.End; i++ {
			sanitizedContent[i] = ' '
		}
		sanitizedBytes += (gap.End - gap.Start)
	}

	// Generate output filename
	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	name := strings.TrimSuffix(filepath.Base(filePath), ext)
	outPath := filepath.Join(dir, fmt.Sprintf("%s_sanitized%s", name, ext))

	err = os.WriteFile(outPath, sanitizedContent, 0644)
	if err != nil {
		return "", 0, fmt.Errorf("failed to write sanitized file: %w", err)
	}

	return outPath, sanitizedBytes, nil
}
