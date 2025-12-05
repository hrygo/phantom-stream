package core

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// SignatureCleaner specifically targets and removes tracking/signature data
type SignatureCleaner struct {
	FilePath string
	Content  []byte
}

// NewSignatureCleaner creates a new signature cleaner instance
func NewSignatureCleaner(filePath string) (*SignatureCleaner, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return &SignatureCleaner{
		FilePath: filePath,
		Content:  content,
	}, nil
}

// CleanSignature performs precise removal of signature data while preserving structure
func (sc *SignatureCleaner) CleanSignature() (string, error) {
	modifiedContent := make([]byte, len(sc.Content))
	copy(modifiedContent, sc.Content)

	// 1. Find and neutralize suspicious embedded file streams
	embeddedFilePattern := regexp.MustCompile(`/F\s*\(\s*sys_stream\.dat\s*\)`)
	matches := embeddedFilePattern.FindAllStringIndex(string(modifiedContent), -1)

	for _, match := range matches {
		// Find the EF reference to locate the stream
		before := string(modifiedContent[:match[0]])

		// Look backwards to find the FileSpec object
		objPattern := regexp.MustCompile(`(\d+\s+0\s+obj).*` + regexp.QuoteMeta(string(modifiedContent[match[0]:match[1]])))
		objMatches := objPattern.FindStringSubmatch(before + string(modifiedContent[match[0]:]))

		if len(objMatches) > 0 {
			objID := strings.Fields(objMatches[1])[0]

			// Find the stream object
			efPattern := regexp.MustCompile(`/EF\s*<<\s*/F\s+` + regexp.QuoteMeta(objID) + `\s+0\s+R`)
			efMatches := efPattern.FindStringSubmatch(string(modifiedContent))

			if len(efMatches) > 0 {
				// Find the referenced stream object and replace its content
				streamRefPattern := regexp.MustCompile(`/F\s+(\d+)\s+0\s+R`)
				streamRefMatches := streamRefPattern.FindStringSubmatch(efMatches[0])

				if len(streamRefMatches) > 1 {
					streamObjID := streamRefMatches[1]

					// Replace stream content with harmless data
					streamPattern := regexp.MustCompile(
						regexp.QuoteMeta(streamObjID) + `\s+0\s+obj\s*<<.*?>>\s*stream\s*(.*?)\s*endstream`,
					)

					// Replace with empty data but keep the structure
					replacement := streamObjID + ` 0 obj<</Filter/FlateDecode/Length 20/Params<</ModDate(D:20241201000000+00'00')/Size 20>>>stream
x+K2KMK"MU(K)KM(
endstream`

					modifiedContent = streamPattern.ReplaceAll(modifiedContent, []byte(replacement))
				}
			}
		}
	}

	// 2. Preserve all other objects, only modify the suspicious stream content

	// Write the cleaned file
	outputPath := sc.FilePath + "_signature_cleaned"
	err := os.WriteFile(outputPath, modifiedContent, 0644)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// CleanMinimal performs minimal cleaning - just emptying suspicious streams
func (sc *SignatureCleaner) CleanMinimal() (string, error) {
	modifiedContent := make([]byte, len(sc.Content))
	copy(modifiedContent, sc.Content)

	// Find object 72 (the stream containing sys_stream.dat)
	obj72Pattern := regexp.MustCompile(`72\s+0\s+obj\s*<<.*?>>\s*stream\s*(.*?)\s*endstream`)

	// Replace with minimal empty stream
	replacement := `72 0 obj<</Filter/FlateDecode/Length 2>>stream
x
endstream`

	modifiedContent = obj72Pattern.ReplaceAll(modifiedContent, []byte(replacement))

	// Write the cleaned file
	outputPath := sc.FilePath + "_minimal_clean"
	err := os.WriteFile(outputPath, modifiedContent, 0644)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// CleanEmbeddedFileOnly removes only the embedded file, preserving FileSpec
func (sc *SignatureCleaner) CleanEmbeddedFileOnly() (string, error) {
	modifiedContent := make([]byte, len(sc.Content))
	copy(modifiedContent, sc.Content)

	// Remove only the EmbeddedFiles dictionary from Catalog
	catalogPattern := regexp.MustCompile(`/EmbeddedFiles\s*<</Names\[\s*<7379735f73747265616d2e646174>\s*73\s+0\s+R\s*\]>>>>`)
	modifiedContent = catalogPattern.ReplaceAll(modifiedContent, []byte{})

	// Keep all objects (73 and 72) but remove their connection
	// This way the structure stays intact but the embedded file is orphaned

	// Write the cleaned file
	outputPath := sc.FilePath + "_embedded_removed"
	err := os.WriteFile(outputPath, modifiedContent, 0644)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// VerifyFileIntegrity checks if the PDF structure is still valid
func VerifyFileIntegrity(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cannot read file: %v", err)
	}

	// Check for basic PDF structure
	if !strings.Contains(string(content), "%PDF-") {
		return fmt.Errorf("missing PDF header")
	}

	if !strings.Contains(string(content), "%%EOF") {
		return fmt.Errorf("missing EOF marker")
	}

	// Count objects
	objPattern := regexp.MustCompile(`\d+\s+\d+\s+obj`)
	objMatches := objPattern.FindAllString(string(content), -1)
	if len(objMatches) < 10 {
		return fmt.Errorf("too few PDF objects (%d), structure may be damaged", len(objMatches))
	}

	// Check for root reference
	trailerPattern := regexp.MustCompile(`trailer\s*<<.*?/Root\s+(\d+\s+\d+\s+R)`)
	trailerMatches := trailerPattern.FindStringSubmatch(string(content))
	if len(trailerMatches) == 0 {
		return fmt.Errorf("missing or invalid trailer")
	}

	return nil
}
