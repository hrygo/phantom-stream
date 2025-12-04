package core

import (
	"bytes"
	"os"
	"regexp"
)

// AdvancedScanResult holds details about structural anomalies
type AdvancedScanResult struct {
	GapAnomalies    int
	HiddenComments  int
	SuspiciousBytes int64
}

// ScanStructure performs a heuristic scan for data hidden between PDF objects.
func ScanStructure(filePath string) (*AdvancedScanResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	res := &AdvancedScanResult{}

	// 1. Comment Analysis
	// Attackers might hide data in comments starting with %.
	// We look for comments that are unusually long or contain high entropy (not implemented here, just counting).
	// A normal comment is usually short.
	// commentCount := bytes.Count(content, []byte("%"))
	// This is too noisy, valid PDFs have many comments.
	// Let's look for "Suspicious Comments" -> Comments that contain non-printable characters?
	// For now, we skip deep comment analysis to avoid false positives.

	// 2. Gap Analysis (Simplified)
	// We look for the pattern: endobj <GAP> <digit> <digit> obj
	// The GAP should be whitespace.

	// Regex to find object boundaries.
	// Note: This is a heuristic and might be fooled by strings containing these keywords.
	// A full parser is needed for 100% accuracy, but this catches lazy steganography.

	// Find all 'endobj' offsets
	endObjMarker := []byte("endobj")
	objMarkerRegex := regexp.MustCompile(`\d+\s+\d+\s+obj`)

	endObjIndices := findAllIndices(content, endObjMarker)

	for _, endIdx := range endObjIndices {
		// Start looking after 'endobj'
		startSearch := endIdx + len(endObjMarker)
		if startSearch >= len(content) {
			continue
		}

		// Find the next 'obj'
		nextObjLoc := objMarkerRegex.FindIndex(content[startSearch:])
		if nextObjLoc == nil {
			// No more objects? Check if we are near xref or trailer
			continue
		}

		gapStart := startSearch
		gapEnd := startSearch + nextObjLoc[0]

		gap := content[gapStart:gapEnd]

		// Check if gap contains non-whitespace
		if containsNonWhitespace(gap) {
			res.GapAnomalies++
			res.SuspiciousBytes += int64(len(gap))
			// fmt.Printf("DEBUG: Found suspicious gap at offset %d, length %d\n", gapStart, len(gap))
		}
	}

	return res, nil
}

func findAllIndices(data, sep []byte) []int {
	var indices []int
	i := 0
	for {
		idx := bytes.Index(data[i:], sep)
		if idx == -1 {
			break
		}
		indices = append(indices, i+idx)
		i += idx + len(sep)
	}
	return indices
}

func containsNonWhitespace(data []byte) bool {
	for _, b := range data {
		switch b {
		case ' ', '\t', '\r', '\n', 0x0C: // 0x0C is Form Feed
			continue
		default:
			return true
		}
	}
	return false
}
