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

		// SAFETY CHECK: Do not prune system objects that might not be referenced via standard "R" links
		// e.g. Object Streams (ObjStm) are referenced by XRef, not by other objects.
		// XRef streams are also special.
		objData := content[node.Offset : node.Offset+node.Length]
		if isSystemObject(objData) {
			// fmt.Printf("Skipping system object %d %d\n", node.ID, node.Gen)
			continue
		}

		// SAFETY CHECK 2: Heuristic Whitelist
		// Since we cannot parse Object Streams (compressed references), our graph is incomplete.
		// We must assume any object that LOOKS like a valid structural object is referenced from somewhere we can't see.
		// We only prune objects that look like "raw data containers" (no /Type, no /Kids, etc).
		if isLikelyValidStructure(objData) {
			// fmt.Printf("Skipping likely valid object %d %d\n", node.ID, node.Gen)
			continue
		}

		// SAFETY CHECK 3: Skip Streams
		// Content streams (pages, images) are often top-level objects referenced by Page objects inside ObjStms.
		// Since we can't see the Page objects, we can't see the links to these streams.
		// Pruning them breaks the document.
		// We conservatively skip ALL streams.
		if strings.Contains(string(objData), "stream") {
			continue
		}

		// Construct a null object replacement
		nullObj := fmt.Sprintf("%d %d obj\nnull\nendobj", node.ID, node.Gen)
		nullBytes := []byte(nullObj)

		if int64(len(nullBytes)) > node.Length {
			// If the original object is smaller than our null replacement, we can't safely replace it in-place
			// without shifting offsets, which breaks XREF.
			// In this rare case, we just blank it out with spaces as before.
			for i := node.Offset; i < node.Offset+node.Length; i++ {
				sanitizedContent[i] = ' '
			}
		} else {
			// Write the null object
			for i := 0; i < len(nullBytes); i++ {
				sanitizedContent[node.Offset+int64(i)] = nullBytes[i]
			}
			// Pad the rest with spaces
			for i := node.Offset + int64(len(nullBytes)); i < node.Offset+node.Length; i++ {
				sanitizedContent[i] = ' '
			}
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

func isSystemObject(data []byte) bool {
	// Check for /Type /ObjStm
	if strings.Contains(string(data), "/ObjStm") {
		return true
	}
	// Check for /Type /XRef
	if strings.Contains(string(data), "/XRef") {
		return true
	}
	// Check for /Type /Metadata (optional, but safer to keep)
	if strings.Contains(string(data), "/Metadata") {
		return true
	}
	// Check for Linearization dict
	if strings.Contains(string(data), "/Linearized") {
		return true
	}
	return false
}

func isLikelyValidStructure(data []byte) bool {
	s := string(data)
	// If it has a /Type, it's likely a valid object (Page, Font, Catalog, etc.)
	if strings.Contains(s, "/Type") {
		return true
	}
	// If it has /Kids (Page Tree)
	if strings.Contains(s, "/Kids") {
		return true
	}
	// If it has /Count (Page Tree)
	if strings.Contains(s, "/Count") {
		return true
	}
	// If it has /Font (Resources)
	if strings.Contains(s, "/Font") {
		return true
	}
	// If it has /ProcSet (Resources)
	if strings.Contains(s, "/ProcSet") {
		return true
	}
	return false
}
