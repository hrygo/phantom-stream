package injector

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// ContentAnchor implements the Phase 8 strategy: Content Stream Perturbation
// It embeds bits by modifying the kerning values in TJ operators
type ContentAnchor struct {
	// Configuration for perturbation
	epsilon float64 // Magnitude of perturbation (e.g., 0.001)
}

// Magic header for content stream payload
var contentMagicHeader = []byte{0xDE, 0xAD, 0xBE, 0xEF}

func NewContentAnchor() *ContentAnchor {
	return &ContentAnchor{
		epsilon: 0.01, // Use 0.01 for better detection
	}
}

func (a *ContentAnchor) Name() string {
	return "Content"
}

// IsAvailable checks if the PDF has pages with content streams
func (a *ContentAnchor) IsAvailable(ctx *model.Context) bool {
	return ctx.PageCount > 0
}

// Inject embeds the payload into page content streams using TJ perturbation
func (a *ContentAnchor) Inject(inputPath, outputPath string, payload []byte) error {
	// Read PDF context
	ctx, err := api.ReadContextFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read context: %w", err)
	}

	// Optimize to ensure all streams are loaded
	if err := api.OptimizeContext(ctx); err != nil {
		return fmt.Errorf("failed to optimize context: %w", err)
	}

	// Prepare payload with magic header
	fullPayload := append(contentMagicHeader, payload...)
	bits := bytesToBits(fullPayload)
	bitIndex := 0

	fmt.Printf("[DEBUG] Content: Payload size %d bytes, %d bits to embed\n", len(fullPayload), len(bits))

	// Find and modify content streams
	injectedBits := 0

	for objNr := 1; objNr <= *ctx.XRefTable.Size; objNr++ {
		entry, found := ctx.Find(objNr)
		if !found || entry.Free || entry.Object == nil {
			continue
		}

		sd, ok := entry.Object.(types.StreamDict)
		if !ok {
			continue
		}

		// Check if this is a content stream (has operators like Tf, TJ, Tj, etc.)
		// Content streams don't have /Subtype like Image XObjects
		if sd.Subtype() != nil {
			continue
		}

		// Decode the stream
		if err := sd.Decode(); err != nil {
			continue
		}

		content := sd.Content
		if len(content) == 0 {
			continue
		}

		// Check if it looks like a content stream (contains text operators)
		contentStr := string(content)
		if !strings.Contains(contentStr, "BT") || !strings.Contains(contentStr, "ET") {
			continue
		}

		fmt.Printf("[DEBUG] Content: Found content stream in object %d (%d bytes)\n", objNr, len(content))

		// Modify TJ operators to embed bits
		modifiedContent, bitsUsed := a.injectBitsIntoContentStream(content, bits, bitIndex)
		if bitsUsed > 0 {
			fmt.Printf("[DEBUG] Content: Injected %d bits into object %d\n", bitsUsed, objNr)
			injectedBits += bitsUsed
			bitIndex += bitsUsed

			// Compress and update the stream
			var buf bytes.Buffer
			w := zlib.NewWriter(&buf)
			w.Write(modifiedContent)
			w.Close()

			sd.Raw = buf.Bytes()
			sd.Content = modifiedContent
			streamLen := int64(len(sd.Raw))
			sd.StreamLength = &streamLen
			sd.InsertName("Filter", "FlateDecode")

			entry.Object = sd
		}

		// Stop if we've embedded all bits
		if bitIndex >= len(bits) {
			break
		}
	}

	if injectedBits == 0 {
		return fmt.Errorf("no TJ operators found for bit injection")
	}

	fmt.Printf("[DEBUG] Content: Total %d/%d bits injected\n", injectedBits, len(bits))

	// Write output
	if err := api.WriteContextFile(ctx, outputPath); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// injectBitsIntoContentStream modifies TJ operators to embed bits
// Returns modified content and number of bits used
func (a *ContentAnchor) injectBitsIntoContentStream(content []byte, bits []int, startIdx int) ([]byte, int) {
	// Regex to find TJ operators: [...] TJ
	// We'll look for number patterns within TJ arrays and modify them
	tjPattern := regexp.MustCompile(`\[([^\]]+)\]\s*TJ`)

	contentStr := string(content)
	bitsUsed := 0
	bitIdx := startIdx

	result := tjPattern.ReplaceAllStringFunc(contentStr, func(match string) string {
		if bitIdx >= len(bits) {
			return match
		}

		// Parse the TJ array content
		arrayStart := strings.Index(match, "[")
		arrayEnd := strings.LastIndex(match, "]")
		if arrayStart < 0 || arrayEnd < 0 {
			return match
		}

		arrayContent := match[arrayStart+1 : arrayEnd]

		// Find and modify numbers in the array
		// Numbers in TJ arrays are kerning adjustments (negative = move right)
		numPattern := regexp.MustCompile(`(-?\d+(?:\.\d+)?)`)
		modifiedArray := numPattern.ReplaceAllStringFunc(arrayContent, func(numStr string) string {
			if bitIdx >= len(bits) {
				return numStr
			}

			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return numStr
			}

			// Encode bit by adding epsilon perturbation
			// bit 0: no change or subtract epsilon
			// bit 1: add epsilon
			if bits[bitIdx] == 1 {
				num += a.epsilon
			} else {
				num -= a.epsilon
			}

			bitIdx++
			bitsUsed++

			// Format with precision to preserve the perturbation
			return fmt.Sprintf("%.3f", num)
		})

		return "[" + modifiedArray + "] TJ"
	})

	return []byte(result), bitsUsed
}

// Extract retrieves the payload from content streams
func (a *ContentAnchor) Extract(filePath string) ([]byte, error) {
	ctx, err := api.ReadContextFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read context: %w", err)
	}

	if err := api.OptimizeContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to optimize context: %w", err)
	}

	var extractedBits []int

	for objNr := 1; objNr <= *ctx.XRefTable.Size; objNr++ {
		entry, found := ctx.Find(objNr)
		if !found || entry.Free || entry.Object == nil {
			continue
		}

		sd, ok := entry.Object.(types.StreamDict)
		if !ok {
			continue
		}

		if sd.Subtype() != nil {
			continue
		}

		// Decode the stream
		if err := sd.Decode(); err != nil {
			continue
		}

		content := sd.Content
		if len(content) == 0 {
			continue
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "BT") || !strings.Contains(contentStr, "ET") {
			continue
		}

		fmt.Printf("[DEBUG] Content: Extracting from object %d\n", objNr)

		// Extract bits from TJ operators
		bits := a.extractBitsFromContentStream(content)
		extractedBits = append(extractedBits, bits...)
	}

	if len(extractedBits) < len(contentMagicHeader)*8 {
		return nil, fmt.Errorf("not enough bits extracted: %d", len(extractedBits))
	}

	// Convert bits to bytes
	data := bitsToBytes(extractedBits)

	// Verify magic header
	if len(data) < len(contentMagicHeader) || !bytes.Equal(data[:len(contentMagicHeader)], contentMagicHeader) {
		return nil, fmt.Errorf("magic header mismatch")
	}

	// Return payload without magic header
	return data[len(contentMagicHeader):], nil
}

// extractBitsFromContentStream extracts bits from TJ operators
func (a *ContentAnchor) extractBitsFromContentStream(content []byte) []int {
	var bits []int

	tjPattern := regexp.MustCompile(`\[([^\]]+)\]\s*TJ`)
	contentStr := string(content)

	matches := tjPattern.FindAllStringSubmatch(contentStr, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		arrayContent := match[1]
		numPattern := regexp.MustCompile(`(-?\d+(?:\.\d+)?)`)
		nums := numPattern.FindAllString(arrayContent, -1)

		for _, numStr := range nums {
			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				continue
			}

			// Decode bit based on fractional part
			// If fractional part indicates positive perturbation -> 1
			// Otherwise -> 0
			frac := num - float64(int(num))
			if frac > 0 {
				bits = append(bits, 1)
			} else {
				bits = append(bits, 0)
			}
		}
	}

	return bits
}

// Helper: Convert bytes to bits
func bytesToBits(data []byte) []int {
	bits := make([]int, 0, len(data)*8)
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			if (b>>i)&1 == 1 {
				bits = append(bits, 1)
			} else {
				bits = append(bits, 0)
			}
		}
	}
	return bits
}

// Helper: Convert bits to bytes
func bitsToBytes(bits []int) []byte {
	if len(bits) == 0 {
		return nil
	}
	numBytes := (len(bits) + 7) / 8
	data := make([]byte, numBytes)
	for i, bit := range bits {
		if bit == 1 {
			byteIdx := i / 8
			bitIdx := 7 - (i % 8)
			data[byteIdx] |= 1 << bitIdx
		}
	}
	return data
}

// Unused import placeholders (will be used in full implementation)
var _ = io.EOF
