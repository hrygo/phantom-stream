package injector

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// ContentAnchor implements the Phase 8/9 strategy: Content Stream Perturbation
// It embeds signature by injecting invisible text with specific kerning values (TJ operator).
type ContentAnchor struct {
}

// Magic header for content stream payload
var contentMagicHeader = []byte{0xDE, 0xAD, 0xBE, 0xEF}

func NewContentAnchor() *ContentAnchor {
	return &ContentAnchor{}
}

func (a *ContentAnchor) Name() string {
	return "Content"
}

// IsAvailable checks if the PDF has pages with content streams
func (a *ContentAnchor) IsAvailable(ctx *model.Context) bool {
	return ctx.PageCount > 0
}

// Inject embeds the payload into page content streams using TJ operator
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

	fmt.Printf("[DEBUG] Content: Payload size %d bytes\n", len(fullPayload))

	// Find and modify content streams
	injectedCount := 0

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

		// Only inject into the first content stream found to avoid duplication
		if injectedCount > 0 {
			continue
		}

		fmt.Printf("[DEBUG] Content: Found content stream in object %d (%d bytes)\n", objNr, len(content))

		// Construct the injection block
		// We use standard text operators but with spaces and kerning values representing our payload.
		// Format:
		// BT
		// /Helv 1 Tf   (Use a standard font, assuming it exists or is default)
		// 3 Tr         (Render mode 3: Neither fill nor stroke -> Invisible)
		// [ ( ) b1 ( ) b2 ... ] TJ
		// ET

		var sb strings.Builder
		sb.WriteString("\nBT\n/Helv 1 Tf\n3 Tr\n[")

		for _, b := range fullPayload {
			// We encode each byte as a kerning value.
			// To avoid confusion with normal text, we can offset it or just use it directly.
			// Using ( ) <val> ensures we have a pattern to match.
			sb.WriteString(fmt.Sprintf(" ( ) %d", b))
		}

		sb.WriteString(" ] TJ\nET\n")
		injectionBlock := []byte(sb.String())

		// Append to the end of the stream
		modifiedContent := append(content, injectionBlock...)

		fmt.Printf("[DEBUG] Content: Appended %d bytes injection block\n", len(injectionBlock))
		injectedCount++

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

	if injectedCount == 0 {
		return fmt.Errorf("no content streams found for injection")
	}

	fmt.Printf("[DEBUG] Content: Injected into %d stream(s)\n", injectedCount)

	// Write output
	if err := api.WriteContextFile(ctx, outputPath); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
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

	// Regex to find our TJ block: \[ ( \( \) \d+ )+ \] TJ
	// Simplified: Look for brackets containing ( ) and numbers
	// We look for the sequence that matches our encoding pattern

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

		// Naive parsing: Look for the specific pattern we injected
		// "3 Tr\n["
		if idx := strings.LastIndex(contentStr, "3 Tr"); idx != -1 {
			// Look ahead for [
			startBracket := strings.Index(contentStr[idx:], "[")
			if startBracket != -1 {
				startBracket += idx
				endBracket := strings.Index(contentStr[startBracket:], "] TJ")
				if endBracket != -1 {
					endBracket += startBracket

					arrayContent := contentStr[startBracket+1 : endBracket]
					// Parse values: ( ) 123 ( ) 456 ...
					// We just want the numbers

					// Split by space
					parts := strings.Fields(arrayContent)
					var extractedBytes []byte

					for i := 0; i < len(parts); i++ {
						part := parts[i]
						if part == "(" || part == ")" || part == "()" {
							continue
						}

						// Try to parse as number
						if val, err := strconv.Atoi(part); err == nil {
							if val >= 0 && val <= 255 {
								extractedBytes = append(extractedBytes, byte(val))
							}
						}
					}

					// Check magic header
					if len(extractedBytes) >= len(contentMagicHeader) {
						// Search for magic header in the extracted bytes
						// Because we might have picked up other numbers
						for i := 0; i <= len(extractedBytes)-len(contentMagicHeader); i++ {
							if bytes.Equal(extractedBytes[i:i+len(contentMagicHeader)], contentMagicHeader) {
								fmt.Printf("[DEBUG] Content: Found payload in object %d\n", objNr)
								return extractedBytes[i+len(contentMagicHeader):], nil
							}
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("content payload not found")
}

// Unused import placeholders
var _ = io.EOF
