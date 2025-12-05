package core

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
)

// HeuristicClean implements dynamic content-hash-based watermark detection and removal.
// It handles cases where the watermark object is duplicated per page (different IDs, same content).
func HeuristicClean(filePath string, threshold float64) (string, int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", 0, err
	}

	// 1. Count Pages (to calculate threshold)
	pageRegex := regexp.MustCompile(`(\d+)\s+0\s+obj[\s\S]*?/Type\s*/Page[\s\S]*?endobj`)
	pageMatches := pageRegex.FindAllIndex(data, -1)
	totalPages := len(pageMatches)

	if totalPages == 0 {
		// Fallback if regex fails or encrypted: assume at least 1
		totalPages = 1
	}

	minCount := int(float64(totalPages) * threshold)
	if minCount < 2 {
		minCount = 2 // Safety floor
	}

	fmt.Printf("[*] Analysis: %d pages found. Threshold: %d repeats.\n", totalPages, minCount)

	// 2. Scan ALL Objects and Hash their Streams
	// Regex: ID 0 obj ... stream ... endstream ... endobj
	// We iterate all objects to find streams
	objRegex := regexp.MustCompile(`(\d+)\s+0\s+obj([\s\S]*?)endobj`)
	streamRegex := regexp.MustCompile(`stream[\r\n\s]+([\s\S]*?)[\r\n\s]+endstream`)

	matches := objRegex.FindAllSubmatchIndex(data, -1)

	type StreamInfo struct {
		ID         string
		RawContent []byte
		Hash       string
		Offset     int // Start of stream content
		End        int // End of stream content
	}

	var streams []StreamInfo
	hashMap := make(map[string]int)

	fmt.Printf("[*] Scanning %d objects for repeated content...\n", len(matches))

	for _, loc := range matches {
		objID := string(data[loc[2]:loc[3]])
		bodyStart := loc[4]
		bodyEnd := loc[5]
		body := data[bodyStart:bodyEnd]

		streamLoc := streamRegex.FindSubmatchIndex(body)
		if streamLoc == nil {
			continue
		}

		streamContentStart := bodyStart + streamLoc[2]
		streamContentEnd := bodyStart + streamLoc[3]
		streamRaw := data[streamContentStart:streamContentEnd]

		// Decompress for accurate hashing (ignore compression differences)
		var contentForHash []byte
		if bytes.Contains(body, []byte("/FlateDecode")) {
			r, err := zlib.NewReader(bytes.NewReader(streamRaw))
			if err == nil {
				contentForHash, _ = io.ReadAll(r)
				r.Close()
			} else {
				contentForHash = streamRaw
			}
		} else {
			contentForHash = streamRaw
		}

		// Skip very small streams (e.g. metadata, simple colors) to avoid false positives
		if len(contentForHash) < 50 {
			continue
		}

		hash := sha256.Sum256(contentForHash)
		hashStr := hex.EncodeToString(hash[:])

		streams = append(streams, StreamInfo{
			ID:         objID,
			RawContent: streamRaw,
			Hash:       hashStr,
			Offset:     streamContentStart,
			End:        streamContentEnd,
		})

		hashMap[hashStr]++
	}

	// 3. Identify Suspicious Hashes
	suspiciousHashes := make(map[string]bool)
	for hash, count := range hashMap {
		if count >= minCount {
			suspiciousHashes[hash] = true
			fmt.Printf("[!] Found Suspicious Content (Hash: %s...): Repeated %d times\n", hash[:8], count)
		}
	}

	if len(suspiciousHashes) == 0 {
		return "", 0, fmt.Errorf("no high-frequency repeated content found")
	}

	// 4. Clean
	modifiedData := make([]byte, len(data))
	copy(modifiedData, data)
	cleanedCount := 0

	for _, s := range streams {
		if suspiciousHashes[s.Hash] {
			// Neutralize
			for i := s.Offset; i < s.End; i++ {
				modifiedData[i] = ' '
			}
			cleanedCount++
		}
	}

	outPath := filePath + "_heuristic_cleaned.pdf"
	err = os.WriteFile(outPath, modifiedData, 0644)
	if err != nil {
		return "", 0, err
	}

	return outPath, cleanedCount, nil
}
