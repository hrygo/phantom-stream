package injector

import (
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

const (
	// AnchorNameVisual is the name of the visual watermark anchor
	AnchorNameVisual = "Visual"
)

// VisualAnchor implements the Phase 9 strategy: Visual Watermarks
// It adds a visible watermark to the PDF pages to deter leaks and increase cleaning cost.
type VisualAnchor struct {
}

func NewVisualAnchor() *VisualAnchor {
	return &VisualAnchor{}
}

func (a *VisualAnchor) Name() string {
	return AnchorNameVisual
}

// IsAvailable checks if the PDF has pages
func (a *VisualAnchor) IsAvailable(ctx *model.Context) bool {
	return ctx.PageCount > 0
}

// Inject adds a visible watermark to the PDF
// Supports full Unicode character range including CJK, Arabic, Cyrillic, etc.
func (a *VisualAnchor) Inject(inputPath, outputPath string, payload []byte) error {
	// Use plaintext payload as watermark content (deterrence, no encryption)
	message := string(payload)

	// Create a watermark configuration
	// Display "CONFIDENTIAL" and the plaintext message
	// Use vertical bar separator for better compatibility
	watermarkText := fmt.Sprintf("CONFIDENTIAL | %s", message)

	// Detect if message contains non-ASCII characters (Unicode)
	isASCII := true
	for _, r := range watermarkText {
		if r > 127 {
			isASCII = false
			break
		}
	}

	var wmConf *model.Watermark
	var err error

	if isASCII {
		// Optimization: Use standard PDF font (Helvetica) for ASCII-only text.
		// This avoids embedding the ~1MB Unicode font, resulting in zero file size overhead.
		wmConf, err = api.TextWatermark(watermarkText,
			"font:Helvetica, points:48, rot:45, op:0.3, col:0.5 0.5 0.5",
			true, false, types.POINTS)
		if err != nil {
			return fmt.Errorf("failed to configure ASCII watermark: %w", err)
		}
	} else {
		// Non-ASCII characters present: Use Image Rasterization optimization.
		// Instead of embedding the full ~14MB (compressed ~1MB) Unicode font,
		// we render the text to a small transparent PNG on the fly and inject that.
		// Overhead becomes negligible (< 50KB).

		// 1. Render text to PNG
		var pngBytes []byte
		var renderErr error
		pngBytes, renderErr = renderTextToPNG(watermarkText)
		if renderErr != nil {
			return fmt.Errorf("failed to render non-ASCII watermark to image: %w", renderErr)
		}

		// 2. Create temp file for the image
		tmpFile, tmpErr := os.CreateTemp("", "phantom_wm_*.png")
		if tmpErr != nil {
			return fmt.Errorf("failed to create temp image file: %w", tmpErr)
		}
		imgParams := "rot:45, op:0.3, scale:1.0 abs" // 1.0 absolute scale (1px = 1pt)

		defer func() {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
		}()

		if _, writeErr := tmpFile.Write(pngBytes); writeErr != nil {
			return fmt.Errorf("failed to write temp image file: %w", writeErr)
		}
		// Close explicitly before using in pdfcpu
		tmpFile.Close()

		// 3. Create Image Watermark configuration
		wmConf, err = api.ImageWatermark(tmpFile.Name(), imgParams, true, false, types.POINTS)
		if err != nil {
			return fmt.Errorf("failed to configure image watermark: %w", err)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to configure watermark: %w", err)
	}

	// Apply watermark to all pages
	// We use api.AddWatermarksFile which handles reading and writing
	if err := api.AddWatermarksFile(inputPath, outputPath, nil, wmConf, nil); err != nil {
		return fmt.Errorf("failed to add watermark: %w", err)
	}

	return nil
}

// Extract for VisualAnchor is a no-op or requires OCR (which we don't do).
// In this architecture, VisualAnchor is for deterrence, not primarily for automated extraction via this tool.
// However, to satisfy the interface, we return nil.
// If we wanted to support extraction, we would need to parse the content stream for the watermark text.
func (a *VisualAnchor) Extract(filePath string) ([]byte, error) {
	// Visual watermarks are intended for human verification or OCR.
	// We don't implement extraction here as it's not a hidden channel.
	return nil, fmt.Errorf("visual watermark extraction not supported")
}
