package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RollbackResult holds info about the rollback operation
type RollbackResult struct {
	OriginalSize   int64
	NewSize        int64
	RevisionsFound int
}

// RollbackPDF removes the last incremental update from the PDF.
func RollbackPDF(filePath string) (string, *RollbackResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Find all %%EOF markers
	marker := []byte("%%EOF")
	indices := findAllIndices(content, marker)

	if len(indices) < 2 {
		return "", nil, fmt.Errorf("no previous revision found (only 1 %%EOF)")
	}

	// We want to keep everything up to the second-to-last %%EOF.
	// The last one is at indices[len-1].
	// The previous one is at indices[len-2].

	// We need to include the marker itself.
	cutoff := indices[len(indices)-2] + len(marker)

	// Check for trailing newline after that marker to be nice
	if cutoff < len(content) {
		if content[cutoff] == '\r' {
			cutoff++
			if cutoff < len(content) && content[cutoff] == '\n' {
				cutoff++
			}
		} else if content[cutoff] == '\n' {
			cutoff++
		}
	}

	newContent := content[:cutoff]

	// Generate output filename
	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	name := strings.TrimSuffix(filepath.Base(filePath), ext)
	outPath := filepath.Join(dir, fmt.Sprintf("%s_rollback%s", name, ext))

	err = os.WriteFile(outPath, newContent, 0644)
	if err != nil {
		return "", nil, fmt.Errorf("failed to write rollback file: %w", err)
	}

	return outPath, &RollbackResult{
		OriginalSize:   int64(len(content)),
		NewSize:        int64(len(newContent)),
		RevisionsFound: len(indices),
	}, nil
}
