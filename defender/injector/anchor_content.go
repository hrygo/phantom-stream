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
// It embeds signature by appending a hidden marked content block at the end of content stream
type ContentAnchor struct {
	// Magic marker for identifying our content
	markerTag string
}

// Magic header for content stream payload
var contentMagicHeader = []byte{0xDE, 0xAD, 0xBE, 0xEF}

func NewContentAnchor() *ContentAnchor {
	return &ContentAnchor{
		markerTag: "DFNDR", // Defender marker tag
	}
}

func (a *ContentAnchor) Name() string {
	return "Content"
}

// IsAvailable checks if the PDF has pages with content streams
func (a *ContentAnchor) IsAvailable(ctx *model.Context) bool {
	return ctx.PageCount > 0
}

// Inject embeds the payload into page content streams using invisible marked content
// This approach appends a hidden comment/marked content block at the end of the stream
// The payload is hex-encoded and wrapped in a graphics state that renders nothing
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
	hexPayload := bytesToHex(fullPayload)

	fmt.Printf("[DEBUG] Content: Payload size %d bytes, hex: %d chars\n", len(fullPayload), len(hexPayload))

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

		// Only inject into the first content stream found
		if injectedCount > 0 {
			continue
		}

		fmt.Printf("[DEBUG] Content: Found content stream in object %d (%d bytes)\n", objNr, len(content))

		// Create invisible marked content block with payload
		// This uses a graphics state that sets text to 100% transparent
		// The payload is embedded as a "hidden" text that renders as invisible
		// Format: q 0 Tr 0 0 0 rg BT /F1 0.001 Tf (HEX_PAYLOAD) Tj ET Q
		// This renders nothing visible but embeds the data in the content stream

		// Simpler approach: use PDF comment syntax which is ignored by renderers
		// % is comment in PostScript/PDF content streams
		// We embed the payload as: \n% DFNDR:HEXDATA\n

		markerBlock := fmt.Sprintf("\n%% %s:%s\n", a.markerTag, hexPayload)
		modifiedContent := append(content, []byte(markerBlock)...)

		fmt.Printf("[DEBUG] Content: Appended %d bytes marker block\n", len(markerBlock))
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
		return fmt.Errorf("failed to write output: %w\n", err)
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

	// Search pattern for our marker
	markerPattern := regexp.MustCompile(`% ` + a.markerTag + `:([0-9A-Fa-f]+)`)

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

		fmt.Fprintf(io.Discard, "[DEBUG] Content: Extracting from object %d\n", objNr)

		// Look for our marker
		matches := markerPattern.FindStringSubmatch(contentStr)
		if len(matches) >= 2 {
			hexData := matches[1]
			data := hexToBytes(hexData)

			// Verify magic header
			if len(data) >= len(contentMagicHeader) && bytes.Equal(data[:len(contentMagicHeader)], contentMagicHeader) {
				fmt.Printf("[DEBUG] Content: Found payload in object %d (%d bytes)\n", objNr, len(data))
				return data[len(contentMagicHeader):], nil
			}
		}
	}

	return nil, fmt.Errorf("content payload not found")
}

// bytesToHex converts bytes to hex string
func bytesToHex(data []byte) string {
	var sb strings.Builder
	for _, b := range data {
		sb.WriteString(fmt.Sprintf("%02X", b))
	}
	return sb.String()
}

// hexToBytes converts hex string to bytes
func hexToBytes(hex string) []byte {
	if len(hex)%2 != 0 {
		return nil
	}
	data := make([]byte, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		b, err := strconv.ParseUint(hex[i:i+2], 16, 8)
		if err != nil {
			return nil
		}
		data[i/2] = byte(b)
	}
	return data
}

// Unused import placeholders (will be used in full implementation)
var _ = io.EOF
