package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PruneZombies removes unreferenced objects from the PDF by overwriting them with spaces.
func PruneZombies(filePath string) (string, int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read file: %w", err)
	}

	graph, err := AnalyzeObjectGraph(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to analyze graph: %w", err)
	}

	if len(graph.ZombieObjects) == 0 {
		return "", 0, nil
	}

	sanitizedContent := make([]byte, len(content))
	copy(sanitizedContent, content)

	var prunedBytes int64

	for _, zombieID := range graph.ZombieObjects {
		node := graph.ObjectMap[zombieID]
		// Overwrite the object body with spaces
		// We keep the "ID Gen obj" and "endobj" markers?
		// No, if we want to be thorough, we blank the whole thing.
		// BUT, if we blank the "obj" keyword, some parsers might get confused if they expect an object there based on XREF.
		// However, since it's unreferenced, the XREF entry *should* be free or pointing to it but nobody uses it.
		// Safest bet: Blank the *content* between "obj" and "endobj".
		// Aggressive bet: Blank everything from start to end.

		// Let's go with Aggressive: Blank everything.
		// If the XREF points to it, it will point to a bunch of spaces.
		// This is valid PDF (whitespace is ignored).

		for i := node.Offset; i < node.Offset+node.Length; i++ {
			sanitizedContent[i] = ' '
		}
		prunedBytes += node.Length
	}

	// Generate output filename
	dir := filepath.Dir(filePath)
	ext := filepath.Ext(filePath)
	name := strings.TrimSuffix(filepath.Base(filePath), ext)
	outPath := filepath.Join(dir, fmt.Sprintf("%s_pruned%s", name, ext))

	err = os.WriteFile(outPath, sanitizedContent, 0644)
	if err != nil {
		return "", 0, fmt.Errorf("failed to write pruned file: %w", err)
	}

	return outPath, len(graph.ZombieObjects), nil
}
