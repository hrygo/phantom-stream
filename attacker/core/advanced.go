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

// GapInterval represents a byte range containing suspicious data
type GapInterval struct {
	Start int64
	End   int64
}

// FindSuspiciousGaps locates gaps between objects that contain non-whitespace data.
func FindSuspiciousGaps(content []byte) ([]GapInterval, error) {
	var gaps []GapInterval

	// Regex to find object boundaries.
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
			continue
		}

		gapStart := startSearch
		gapEnd := startSearch + nextObjLoc[0]

		gap := content[gapStart:gapEnd]

		// Check if gap contains non-whitespace
		if containsNonWhitespace(gap) {
			gaps = append(gaps, GapInterval{
				Start: int64(gapStart),
				End:   int64(gapEnd),
			})
		}
	}
	return gaps, nil
}

// ScanStructure performs a heuristic scan for data hidden between PDF objects.
func ScanStructure(filePath string) (*AdvancedScanResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	res := &AdvancedScanResult{}

	// 1. Comment Analysis (Skipped for now)
	// commentCount := bytes.Count(content, []byte("%"))

	// 2. Gap Analysis
	gaps, err := FindSuspiciousGaps(content)
	if err != nil {
		return nil, err
	}

	res.GapAnomalies = len(gaps)
	for _, g := range gaps {
		res.SuspiciousBytes += (g.End - g.Start)
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
