package injector

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// VisualAnchor implements the Phase 9 strategy: Visual Watermarks
// It adds a visible watermark to the PDF pages to deter leaks and increase cleaning cost.
type VisualAnchor struct {
}

func NewVisualAnchor() *VisualAnchor {
	return &VisualAnchor{}
}

func (a *VisualAnchor) Name() string {
	return "Visual"
}

// IsAvailable checks if the PDF has pages
func (a *VisualAnchor) IsAvailable(ctx *model.Context) bool {
	return ctx.PageCount > 0
}

// Inject adds a visible watermark to the PDF
func (a *VisualAnchor) Inject(inputPath, outputPath string, payload []byte) error {
	// Use plaintext payload as watermark content (deterrence, no encryption)
	message := string(payload)

	// Create a watermark configuration
	// Display "CONFIDENTIAL" and the plaintext message
	watermarkText := fmt.Sprintf("CONFIDENTIAL\n%s", message)

	// Configure watermark
	// Rotation: 45 degrees, Opacity: 0.3, Font: Helvetica, Size: 48, Color: Gray
	wmConf, err := api.TextWatermark(watermarkText, "rot:45, op:0.3, font:Helvetica, points:48, col:0.5 0.5 0.5", true, false, types.POINTS)
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
