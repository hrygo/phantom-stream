package core

import (
	"os"
	"regexp"
	"strconv" // Added import
)

// RemoveSuspiciousContentOnly removes only the suspicious stream content, preserving structure
func RemoveSuspiciousContentOnly(filePath string) (string, *SemanticResult, error) {
	// First analyze the file
	result, err := AnalyzeEmbeddedFiles(filePath)
	if err != nil {
		return "", nil, err
	}

	if len(result.SuspiciousFiles) == 0 {
		return "", result, nil
	}

	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil, err
	}

	// Create output path
	outPath := filePath + "_conservative"

	// Conservatively remove only stream content, preserve structure
	modifiedContent := content

	for _, file := range result.SuspiciousFiles {
		// Find and clear only the stream content of suspicious embedded files
		// Look for the stream object and replace its content with placeholder
		if file.ObjectID > 0 {
			// Fix: Use strconv.Itoa to convert int to string
			objPattern := regexp.MustCompile(regexp.QuoteMeta(strconv.Itoa(file.ObjectID)) + `\s+0\s+obj\s*<</EF<</F\s+(\d+)\s+0\s+R`)
			objMatches := objPattern.FindStringSubmatch(string(modifiedContent))

			if len(objMatches) > 1 {
				streamObjID := objMatches[1]

				// Replace stream content with empty placeholder
				streamPattern := regexp.MustCompile(regexp.QuoteMeta(streamObjID) + `\s+0\s+obj\s*<<.*?>>\s*stream\s*.*?\s*endstream`)
				placeholder := streamObjID + ` 0 obj<</Length 0>>stream
endstream`
				modifiedContent = streamPattern.ReplaceAll(modifiedContent, []byte(placeholder))
			}
		}
	}

	// Write conservatively modified file
	err = os.WriteFile(outPath, modifiedContent, 0644)
	if err != nil {
		return "", nil, err
	}

	return outPath, result, nil
}
