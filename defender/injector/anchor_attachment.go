package injector

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// AttachmentAnchor implements signature embedding via PDF attachments
type AttachmentAnchor struct{}

// NewAttachmentAnchor creates a new attachment anchor
func NewAttachmentAnchor() *AttachmentAnchor {
	return &AttachmentAnchor{}
}

// Name returns the anchor type name
func (a *AttachmentAnchor) Name() string {
	return "Attachment"
}

// Inject embeds the payload as a PDF attachment
func (a *AttachmentAnchor) Inject(inputPath, outputPath string, payload []byte) error {
	// Create temporary directory for attachment
	tmpDir, err := os.MkdirTemp("", "defender_attach_*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if cleanErr := os.RemoveAll(tmpDir); cleanErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean temp directory: %v\n", cleanErr)
		}
	}()

	// Write payload to temporary file
	payloadPath := filepath.Join(tmpDir, attachName)
	if err := os.WriteFile(payloadPath, payload, 0600); err != nil {
		return fmt.Errorf("failed to write payload: %w", err)
	}

	// Add attachment to PDF
	conf := model.NewDefaultConfiguration()
	if err := api.AddAttachmentsFile(inputPath, outputPath, []string{payloadPath}, true, conf); err != nil {
		return fmt.Errorf("failed to add attachment to PDF: %w", err)
	}

	return nil
}

// Extract retrieves the payload from PDF attachment
func (a *AttachmentAnchor) Extract(filePath string) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "defender_verify_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if cleanErr := os.RemoveAll(tmpDir); cleanErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clean temp directory: %v\n", cleanErr)
		}
	}()

	conf := model.NewDefaultConfiguration()

	// Extract attachment from PDF
	if err := api.ExtractAttachmentsFile(filePath, tmpDir, []string{attachName}, conf); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAttachmentNotFound, err)
	}

	extractedPath := filepath.Join(tmpDir, attachName)
	payload, err := os.ReadFile(extractedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read extracted attachment: %w", err)
	}

	return payload, nil
}

// IsAvailable checks if attachment anchor can be used
// Attachment anchor is always available (doesn't require special PDF features)
func (a *AttachmentAnchor) IsAvailable(ctx *model.Context) bool {
	return true
}
