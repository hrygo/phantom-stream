package core

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
)

// EmbeddedFile represents an embedded file in PDF
type EmbeddedFile struct {
	ObjectID       int
	Size           int64
	IsCompressed   bool
	Filter         string
	HasParams      bool
	ParamChecksum  string
	ParamModDate   string
	ParamSize      int64
	ContentEntropy float64
	IsSuspicious   bool
	SuspicionScore float64
	Reasons        []string
}

// SemanticResult holds the semantic analysis results
type SemanticResult struct {
	TotalEmbeddedFiles   int
	SuspiciousFiles      []EmbeddedFile
	CleanFiles           []EmbeddedFile
	TotalSuspiciousBytes int64
}

// AnalyzeEmbeddedFiles performs semantic analysis on PDF embedded files
func AnalyzeEmbeddedFiles(filePath string) (*SemanticResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	result := &SemanticResult{
		TotalEmbeddedFiles: 0,
		SuspiciousFiles:    make([]EmbeddedFile, 0),
		CleanFiles:         make([]EmbeddedFile, 0),
	}

	// 1. Find EmbeddedFiles dictionary - allow more flexible patterns
	embeddedFilesPatterns := []string{
		`<</EmbeddedFiles<</Names\[(.+?)\]>>>`,
		`/EmbeddedFiles\s*<<\s*/Names\s*\[\s*(.+?)\s*\]`,
	}

	var namesStr string
	for _, pattern := range embeddedFilesPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindSubmatch(content)
		if len(matches) > 0 {
			namesStr = string(matches[1])
			break
		}
	}

	if namesStr == "" {
		// No embedded files found
		return result, nil
	}

	// 2. Extract file specification references - handle both hex names and direct references
	namePatterns := []string{
		`<([0-9a-fA-F]+)>\s+(\d+)\s+0\s+R>`,
		`(\d+)\s+0\s+R`,
	}

	var nameMatches [][]string
	for _, pattern := range namePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(namesStr, -1)
		if len(matches) > 0 {
			nameMatches = matches
			break
		}
	}

	for _, match := range nameMatches {
		objID := parseInt(match[1])
		fileName := ""
		if len(match) > 2 {
			fileName = hexToString(match[1])
		}

		// 3. Analyze each embedded file object
		embeddedFile, err := analyzeEmbeddedFileObject(content, objID, fileName)
		if err != nil {
			continue
		}

		result.TotalEmbeddedFiles++

		if embeddedFile.IsSuspicious {
			result.SuspiciousFiles = append(result.SuspiciousFiles, embeddedFile)
			result.TotalSuspiciousBytes += embeddedFile.Size
		} else {
			result.CleanFiles = append(result.CleanFiles, embeddedFile)
		}
	}

	return result, nil
}

// analyzeEmbeddedFileObject analyzes a specific embedded file object
func analyzeEmbeddedFileObject(content []byte, objID int, hexFileName string) (EmbeddedFile, error) {
	file := EmbeddedFile{
		ObjectID: objID,
		Reasons:  make([]string, 0),
	}

	// Find the object definition
	objPattern := regexp.MustCompile(fmt.Sprintf(`%d\s+0\s+obj\s*(.*?)\s*endobj`, objID))
	objMatches := objPattern.FindStringSubmatch(string(content))

	if len(objMatches) == 0 {
		return file, fmt.Errorf("object %d not found", objID)
	}

	objContent := objMatches[1]

	// Extract filename from /F entry if hex name is not available
	fileName := hexFileName
	if fileName == "" {
		fPattern := regexp.MustCompile(`/F\(([^)]+)\)`)
		fMatches := fPattern.FindStringSubmatch(objContent)
		if len(fMatches) > 0 {
			fileName = fMatches[1]
			// Clean up zero-width characters
			fileName = strings.ReplaceAll(fileName, "\x00", "")
		}
	}

	// Check for FileSpec structure (should contain /F or /EF)
	if !strings.Contains(objContent, "/F") && !strings.Contains(objContent, "/EF") {
		file.IsSuspicious = true
		file.SuspicionScore += 0.3
		file.Reasons = append(file.Reasons, "Missing standard FileSpec structure")
	}

	// Find the actual embedded file stream (usually in a nested object)
	efPattern := regexp.MustCompile(`/EF\s*<<\s*(?:/F\s+(\d+\s+0\s+R)|/UF\s+(\d+\s+0\s+R))`)
	efMatches := efPattern.FindStringSubmatch(objContent)

	var streamObjID int
	if len(efMatches) > 0 {
		// Extract stream object ID from either /F or /UF
		for i := 1; i < len(efMatches); i++ {
			if efMatches[i] != "" {
				streamParts := strings.Fields(efMatches[i])
				if len(streamParts) >= 1 {
					streamObjID = parseInt(streamParts[0])
					break
				}
			}
		}

		if streamObjID > 0 {
			// Analyze the stream object
			streamSize, isCompressed, filter, entropy := analyzeStreamObject(content, streamObjID)

			file.Size = streamSize
			file.IsCompressed = isCompressed
			file.Filter = filter
			file.ContentEntropy = entropy

			// Also check Params for additional metadata
			paramsPattern := regexp.MustCompile(`/Params<<([^>>]+)>>`)
			paramsMatches := paramsPattern.FindStringSubmatch(objContent)
			if len(paramsMatches) > 0 {
				paramsStr := paramsMatches[1]

				// Extract Size from Params if available
				sizePattern := regexp.MustCompile(`/Size\s+(\d+)`)
				sizeMatches := sizePattern.FindStringSubmatch(paramsStr)
				if len(sizeMatches) > 0 {
					paramSize := parseInt64(sizeMatches[1])
					if paramSize > 0 && (file.Size == 0 || paramSize != file.Size) {
						file.Size = paramSize
					}
				}
			}

			// Apply semantic rules
			applySemanticRules(&file, fileName)
		}
	}

	return file, nil
}

// analyzeStreamObject analyzes a PDF stream object
func analyzeStreamObject(content []byte, objID int) (int64, bool, string, float64) {
	objPattern := regexp.MustCompile(fmt.Sprintf(`%d\s+0\s+obj\s*<<\s*(.*?)\s*/Length\s+(\d+)`, objID))
	objMatches := objPattern.FindStringSubmatch(string(content))

	if len(objMatches) == 0 {
		return 0, false, "", 0
	}

	dictStr := objMatches[1]
	length := parseInt64(objMatches[2])

	// Check for compression filter
	isCompressed := false
	filter := "none"
	if strings.Contains(dictStr, "/Filter") {
		if strings.Contains(dictStr, "/FlateDecode") {
			isCompressed = true
			filter = "FlateDecode"
		} else if strings.Contains(dictStr, "/DCTDecode") {
			isCompressed = true
			filter = "DCTDecode"
		} else if strings.Contains(dictStr, "/CCITTFaxDecode") {
			isCompressed = true
			filter = "CCITTFaxDecode"
		}
	}

	// Extract stream content for entropy calculation
	streamPattern := regexp.MustCompile(fmt.Sprintf(`%d\s+0\s+obj\s*<<.*?>>\s*stream\s*(.*?)\s*endstream`, objID))
	streamMatches := streamPattern.FindStringSubmatch(string(content))

	var entropy float64 = 0
	if len(streamMatches) > 1 {
		streamContent := streamMatches[1]
		entropy = calculateEntropy([]byte(streamContent))
	}

	return length, isCompressed, filter, entropy
}

// applySemanticRules applies semantic analysis rules to detect suspicious embedded files
func applySemanticRules(file *EmbeddedFile, fileName string) {
	// Clean up filename
	cleanName := strings.ReplaceAll(fileName, "\x00", "")
	lowerName := strings.ToLower(cleanName)

	// Rule 1: Check file extension and type
	suspiciousExts := []string{".exe", ".dll", ".bat", ".cmd", ".scr", ".vbs", ".js", ".jar", ".ps1", ".dat"}
	for _, ext := range suspiciousExts {
		if strings.HasSuffix(lowerName, ext) {
			file.IsSuspicious = true
			file.SuspicionScore += 0.6
			file.Reasons = append(file.Reasons, fmt.Sprintf("Suspicious file extension: %s", ext))
		}
	}

	// Rule 2: Check for suspicious filename patterns
	suspiciousPatterns := []string{"secret", "payload", "backdoor", "shell", "sys_", "stream", "temp", "cache"}
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerName, pattern) {
			file.IsSuspicious = true
			file.SuspicionScore += 0.4
			file.Reasons = append(file.Reasons, fmt.Sprintf("Suspicious filename pattern: %s", pattern))
		}
	}

	// Rule 3: Check file size
	if file.Size > 10*1024*1024 { // > 10MB
		file.IsSuspicious = true
		file.SuspicionScore += 0.4
		file.Reasons = append(file.Reasons, "Large embedded file (>10MB)")
	}

	// Rule 4: Check entropy (high entropy may indicate encrypted/compressed payload)
	if file.ContentEntropy > 7.5 && !file.IsCompressed {
		file.IsSuspicious = true
		file.SuspicionScore += 0.5
		file.Reasons = append(file.Reasons, "High entropy in uncompressed content")
	}

	// Rule 5: Check if it's a non-standard attachment type
	if !isStandardAttachmentType(cleanName) && file.Size > 1024 {
		file.IsSuspicious = true
		file.SuspicionScore += 0.3
		file.Reasons = append(file.Reasons, "Non-standard attachment type with significant size")
	}

	// Rule 6: Check for empty or suspicious names
	if cleanName == "" || len(cleanName) < 3 {
		file.IsSuspicious = true
		file.SuspicionScore += 0.5
		file.Reasons = append(file.Reasons, "Empty or very short filename")
	}

	// Rule 7: Special detection for "sys_stream.dat" pattern
	if strings.Contains(lowerName, "sys_stream") && strings.Contains(lowerName, ".dat") {
		file.IsSuspicious = true
		file.SuspicionScore += 0.8
		file.Reasons = append(file.Reasons, "Highly suspicious filename: sys_stream.dat pattern")
	}

	// Determine if file is suspicious based on score
	if file.SuspicionScore >= 0.5 {
		file.IsSuspicious = true
	}
}

// isStandardAttachmentType checks if the file is a standard PDF attachment type
func isStandardAttachmentType(fileName string) bool {
	standardTypes := []string{
		".pdf", ".txt", ".xml", ".html", ".htm", ".css", ".jpg", ".jpeg",
		".png", ".gif", ".tif", ".tiff", ".bmp", ".zip", ".rar", ".7z",
		".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".odt", ".ods", ".odp",
	}

	lowerName := strings.ToLower(fileName)
	for _, ext := range standardTypes {
		if strings.HasSuffix(lowerName, ext) {
			return true
		}
	}
	return false
}

// calculateEntropy calculates the Shannon entropy of data
func calculateEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}

	// Count byte frequencies
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}

	// Calculate entropy
	var entropy float64
	dataLen := float64(len(data))

	for _, count := range freq {
		if count > 0 {
			p := float64(count) / dataLen
			entropy -= p * math.Log2(p)
		}
	}

	return entropy
}

// RemoveSuspiciousAttachments removes suspicious embedded files from PDF
func RemoveSuspiciousAttachments(filePath string) (string, *SemanticResult, error) {
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
	outPath := filePath + "_sanitized"

	// Remove suspicious embedded file objects
	modifiedContent := content
	for _, file := range result.SuspiciousFiles {
		// Remove the embedded file object
		objPattern := regexp.MustCompile(fmt.Sprintf(`%d\s+0\s+obj\s*.*?\s*endobj\s*`, file.ObjectID))
		modifiedContent = objPattern.ReplaceAll(modifiedContent, []byte{})

		// Remove from EmbeddedFiles Names array
		namesPattern := regexp.MustCompile(`<([0-9a-fA-F]+)>\s+%d\s+0\s+R>`)
		modifiedContent = namesPattern.ReplaceAll(modifiedContent, []byte{})
	}

	// Remove the entire EmbeddedFiles structure if no files remain
	if len(result.CleanFiles) == 0 {
		embeddedFilesPattern := regexp.MustCompile(`<</EmbeddedFiles<</Names\[.*?\]>>>>`)
		modifiedContent = embeddedFilesPattern.ReplaceAll(modifiedContent, []byte{})
	}

	// Write sanitized file
	err = os.WriteFile(outPath, modifiedContent, 0644)
	if err != nil {
		return "", nil, err
	}

	return outPath, result, nil
}

// Helper functions
func parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

func parseInt64(s string) int64 {
	var i int64
	fmt.Sscanf(s, "%d", &i)
	return i
}

func hexToString(hexStr string) string {
	var result string
	for i := 0; i < len(hexStr); i += 2 {
		if i+1 < len(hexStr) {
			var b byte
			fmt.Sscanf(hexStr[i:i+2], "%02x", &b)
			result += string(b)
		}
	}
	return result
}