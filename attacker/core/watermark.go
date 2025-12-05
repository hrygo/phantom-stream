package core

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"regexp"
)

// RemoveSpecificWatermark scans for objects containing the specific watermark signature
// and neutralizes them by overwriting the stream content with spaces.
// This preserves file offsets/XRef table.
func RemoveSpecificWatermark(filePath string) (string, int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", 0, err
	}

	// Create a copy to modify
	modifiedData := make([]byte, len(data))
	copy(modifiedData, data)

	// Regex to find objects: ID 0 obj ... endobj
	// We capture the ID and the content body
	objRegex := regexp.MustCompile(`(\d+)\s+0\s+obj([\s\S]*?)endobj`)

	// Regex to find stream inside the object body
	streamRegex := regexp.MustCompile(`stream[\r\n\s]+([\s\S]*?)[\r\n\s]+endstream`)

	matches := objRegex.FindAllSubmatchIndex(data, -1)

	count := 0
	signature := []byte("<b78b") // The specific hex signature of the watermark

	fmt.Printf("[*] Scanning %d objects for watermark signature...\n", len(matches))

	for _, loc := range matches {
		// loc[0]: start of match, loc[1]: end of match
		// loc[2]: start of ID group, loc[3]: end of ID group
		// loc[4]: start of body group, loc[5]: end of body group

		objID := string(data[loc[2]:loc[3]])
		bodyStart := loc[4]
		bodyEnd := loc[5]
		body := data[bodyStart:bodyEnd]

		// Find stream within the body
		streamLoc := streamRegex.FindSubmatchIndex(body)
		if streamLoc == nil {
			continue
		}

		// streamLoc indices are relative to 'body'
		// streamLoc[2]: start of stream content (group 1)
		// streamLoc[3]: end of stream content (group 1)

		streamContentStart := bodyStart + streamLoc[2]
		streamContentEnd := bodyStart + streamLoc[3]
		streamRaw := data[streamContentStart:streamContentEnd]

		// Check if compressed
		isCompressed := bytes.Contains(body, []byte("/FlateDecode"))

		var contentToCheck []byte
		if isCompressed {
			r, err := zlib.NewReader(bytes.NewReader(streamRaw))
			if err == nil {

				decompressed, _ := io.ReadAll(r)
				r.Close()
				contentToCheck = decompressed
			} else {
				// If decompression fails, check raw (unlikely to match but safe)
				contentToCheck = streamRaw
			}
		} else {
			contentToCheck = streamRaw
		}

		// Check for signature
		if bytes.Contains(contentToCheck, signature) {
			fmt.Printf("[!] Found Watermark in Object %s\n", objID)

			// Neutralize: Replace raw stream content with spaces
			// This preserves the file size and offsets
			for i := streamContentStart; i < streamContentEnd; i++ {
				modifiedData[i] = ' '
			}
			count++
		}
	}

	if count == 0 {
		return "", 0, fmt.Errorf("no watermark objects found")
	}

	outPath := filePath + "_watermark_cleaned.pdf"
	err = os.WriteFile(outPath, modifiedData, 0644)
	if err != nil {
		return "", 0, err
	}

	return outPath, count, nil
}
