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

// ComprehensiveClean performs the "Full-Spectrum Cleaning" as described in the Joint Report.
// It combines Attachment wiping, SMask sanitization, and Heuristic content cleaning.
func ComprehensiveClean(filePath string, heuristicThreshold float64) (string, error) {
	fmt.Printf("[*] Starting Full-Spectrum Cleaning on %s...\n", filePath)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// 1. Wipe Attachments (Phase 9 Anchor 1)
	fmt.Println("--- Stage 1: Attachment Neutralization ---")
	content, countAtt := wipeAttachments(content)
	fmt.Printf("[+] Neutralized %d embedded file streams.\n", countAtt)

	// 2. Sanitize SMasks (Phase 9 Anchor 2)
	fmt.Println("--- Stage 2: SMask Sanitization ---")
	content, countSMask := wipeSMasks(content)
	fmt.Printf("[+] Sanitized %d SMask transparency layers.\n", countSMask)

	// 3. Heuristic Cleaning (Phase 9 Visual & Content Anchors)
	fmt.Println("--- Stage 3: Heuristic Stream Cleaning ---")
	// We use a lower threshold for "All-in-One" to be aggressive, or user provided
	content, countHeuristic, err := wipeHeuristic(content, heuristicThreshold)
	if err != nil {
		fmt.Printf("[!] Heuristic scan warning: %v\n", err)
	} else {
		fmt.Printf("[+] Neutralized %d high-frequency repeated streams.\n", countHeuristic)
	}

	// 4. Save
	outPath := filePath + "_final_cleaned.pdf"
	err = os.WriteFile(outPath, content, 0644)
	if err != nil {
		return "", err
	}

	fmt.Printf("[+] Full-Spectrum Cleaning Complete. Saved to: %s\n", outPath)
	return outPath, nil
}

func wipeAttachments(content []byte) ([]byte, int) {
	// Find EmbeddedFiles Name Tree
	// /EmbeddedFiles << /Names [ ... ] >>
	// We look for the /Names array

	// Note: This simple regex might miss complex name trees, but works for the exercise scope.
	// Finding object IDs referenced in the Names array.
	// Pattern: (Name) (Ref)
	// Ref is "123 0 R"

	// First, find the /Names array content inside /EmbeddedFiles
	efRegex := regexp.MustCompile(`/EmbeddedFiles\s*<<\s*/Names\s*\[([\s\S]*?)\]`)
	matches := efRegex.FindSubmatch(content)

	if matches == nil {
		return content, 0
	}

	namesContent := matches[1] // Inside the brackets

	// Find all references "123 0 R"
	refRegex := regexp.MustCompile(`(\d+)\s+0\s+R`)
	refs := refRegex.FindAllSubmatch(namesContent, -1)

	count := 0
	cleanedIDs := make(map[string]bool)

	for _, ref := range refs {
		idStr := string(ref[1])
		if cleanedIDs[idStr] {
			continue
		}

		id := 0
		fmt.Sscanf(idStr, "%d", &id)

		// Wipe the stream of this object
		// We use CleanObjectStreamInContent which handles the safe padding
		newContent, err := CleanObjectStreamInContent(content, id)
		if err == nil {
			content = newContent
			count++
			cleanedIDs[idStr] = true
		}
	}

	return content, count
}

func wipeSMasks(content []byte) ([]byte, int) {
	// Strategy: "Canonicalization" (Lossless Recompression)
	// We do NOT disable the mask (which breaks visuals).
	// We do NOT wipe the mask (which breaks visuals).
	// We Recompress the mask to strip appended data/artifacts.

	smaskIDs := FindSMaskObjects(content)
	fmt.Printf("[*] Found %d SMask object(s) to canonicalize.\n", len(smaskIDs))

	count := 0
	for _, id := range smaskIDs {
		// Use Lossless Canonicalization
		newContent, err := SanitizeImageStreamInContent(content, id)
		if err == nil {
			content = newContent
			count++
		} else {
			fmt.Printf("[!] Warning: Failed to canonicalize SMask %d: %v\n", id, err)
		}
	}
	return content, count
}

func wipeHeuristic(data []byte, threshold float64) ([]byte, int, error) {
	// 1. Count Pages
	pageRegex := regexp.MustCompile(`(\d+)\s+0\s+obj[\s\S]*?/Type\s*/Page[\s\S]*?endobj`)
	pageMatches := pageRegex.FindAllIndex(data, -1)
	totalPages := len(pageMatches)

	if totalPages == 0 {
		totalPages = 1
	}

	minCount := int(float64(totalPages) * threshold)
	if minCount < 2 {
		minCount = 2
	}

	fmt.Printf("[*] Heuristic Threshold: > %d repeats (%.0f%%)\n", minCount, threshold*100)

	// 2. Scan Objects & Hash
	objRegex := regexp.MustCompile(`(\d+)\s+0\s+obj([\s\S]*?)endobj`)
	streamRegex := regexp.MustCompile(`stream[\r\n\s]+([\s\S]*?)[\r\n\s]+endstream`)

	matches := objRegex.FindAllSubmatchIndex(data, -1)

	type StreamInfo struct {
		ID         string
		RawContent []byte // We need raw for replacement location, but hashing uses decoded
		Hash       string
		Offset     int
		End        int
	}

	var streams []StreamInfo
	hashMap := make(map[string]int)

	for _, loc := range matches {
		// objID := string(data[loc[2]:loc[3]])
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

		// Decompress for hashing
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

		if len(contentForHash) < 50 {
			continue
		}

		hash := sha256.Sum256(contentForHash)
		hashStr := hex.EncodeToString(hash[:])

		streams = append(streams, StreamInfo{
			Hash:   hashStr,
			Offset: streamContentStart,
			End:    streamContentEnd,
		})

		hashMap[hashStr]++
	}

	// 3. Identify Suspicious
	suspiciousHashes := make(map[string]bool)
	for hash, count := range hashMap {
		if count >= minCount {
			suspiciousHashes[hash] = true
			fmt.Printf("[!] Found Suspicious Repeated Content (Hash: %s...): %d times\n", hash[:8], count)
		}
	}

	if len(suspiciousHashes) == 0 {
		return data, 0, nil
	}

	// 4. Clean
	// We modify 'data' in place (it's a slice copy passed in, but we return it)
	// Actually, better to work on a copy or carefully since multiple streams might be close?
	// Since we are just replacing with spaces, the length doesn't change, so indices remain valid.
	// BUT if we process in arbitrary order, we're fine as long as ranges don't overlap.
	// Object streams don't overlap.

	count := 0
	for _, s := range streams {
		if suspiciousHashes[s.Hash] {
			for i := s.Offset; i < s.End; i++ {
				data[i] = ' '
			}
			count++
		}
	}

	return data, count, nil
}
