package injector

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// renderTextToPNG renders the given text to a transparent PNG using the embedded Unicode font.
// It returns the PNG bytes or an error.
func renderTextToPNG(text string, fontSize float64) ([]byte, error) {
	if len(goNotoCurrentTTF) == 0 {
		return nil, fmt.Errorf("embedded font data is empty")
	}

	// Parse the embedded font
	f, err := opentype.Parse(goNotoCurrentTTF)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded font: %w", err)
	}

	const (
		dpi = 72.0
	)

	// Create a font face
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingNone,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create font face: %w", err)
	}
	defer face.Close()

	// Measure string to determine image dimensions
	drawer := &font.Drawer{
		Face: face,
	}

	// Measurement
	bounds, advance := drawer.BoundString(text)

	// Convert fixed.Int26_6 to int (ceil)
	// bounds.Max.Y is usually baseline to bottom descent, Min.Y is baseline to top ascent (negative)
	// Height = (Max.Y - Min.Y)
	// Width = advance

	h := (bounds.Max.Y - bounds.Min.Y).Ceil()
	w := advance.Ceil()

	// Add some padding
	padding := 10
	width := w + padding*2
	height := h + padding*2

	// Create RGBA image (transparent background)
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Setup drawer
	drawer.Dst = img
	drawer.Src = image.NewUniform(color.RGBA{128, 128, 128, 77}) // Color: Grey (0.5), Alpha 0.3 (~77/255)
	// Note: api.TextWatermark uses "col:0.5 0.5 0.5, op:0.3".
	// image/draw handles alpha premultiplication.
	// Actually, pdfcpu handles opacity ('op:0.3') separately for images if we pass it in configuration.
	// If we bake alpha into the pixel, and THEN pdfcpu applies opacity, it might be double-faded.
	// Current plan: Generate fully opaque grey text (0.5, 0.5, 0.5) and let pdfcpu handle "op:0.3".
	drawer.Src = image.NewUniform(color.RGBA{128, 128, 128, 255})

	// Set dot position (baseline)
	// Text starts at padding, and baseline is roughly -Min.Y + padding
	baselineY := -bounds.Min.Y + fixed.I(padding)
	drawer.Dot = fixed.Point26_6{
		X: fixed.I(padding),
		Y: baselineY,
	}

	// Draw
	drawer.DrawString(text)

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return buf.Bytes(), nil
}
