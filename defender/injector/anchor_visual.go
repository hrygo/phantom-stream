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
	// Use plaintext message directly as watermark text
	watermarkText := message

	// Adaptive font size calculation
	runes := []rune(watermarkText)
	charCount := len(runes)
	baseSize := 48.0
	targetWidth := 800.0 // Effective diagonal space usually available

	fontSize := baseSize
	if float64(charCount)*baseSize > targetWidth {
		fontSize = targetWidth / float64(charCount)
	}
	// Clamp min size
	if fontSize < 20.0 {
		fontSize = 20.0
	}

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
		desc := fmt.Sprintf("font:Helvetica, points:%.1f, rot:45, op:0.3, col:0.5 0.5 0.5", fontSize)
		wmConf, err = api.TextWatermark(watermarkText,
			desc,
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
		pngBytes, renderErr = renderTextToPNG(watermarkText, fontSize)
		if renderErr != nil {
			return fmt.Errorf("failed to render non-ASCII watermark to image: %w", renderErr)
		}

		// 2. Create temp file for the image
		tmpFile, tmpErr := os.CreateTemp("", "phantom_wm_*.png")
		if tmpErr != nil {
			return fmt.Errorf("failed to create temp image file: %w", tmpErr)
		}
		// Scale is 1.0 abs because the image is already created with the correct pixel size for the font size
		// However, pdfcpu treats image watermark size based on actual image dimensions.
		// If we rendered at 72 DPI, 1 pixel = 1 point.
		// So scale:1.0 abs is correct.
		imgParams := "rot:45, op:0.3, scale:1.0 abs"

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
