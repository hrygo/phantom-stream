package injector

import "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

// Anchor defines the interface for signature embedding mechanisms
// Each anchor type implements a different steganographic technique
type Anchor interface {
	// Name returns the anchor type name (e.g., "Attachment", "SMask")
	Name() string

	// Inject embeds the payload into the PDF
	// Returns error if injection fails
	Inject(inputPath, outputPath string, payload []byte) error

	// Extract retrieves the payload from the PDF
	// Returns the payload bytes or error if extraction fails
	Extract(filePath string) ([]byte, error)

	// IsAvailable checks if this anchor type can be used with the given PDF
	// For example, SMask requires at least one image
	IsAvailable(ctx *model.Context) bool
}

// AnchorRegistry manages available anchor implementations
type AnchorRegistry struct {
	anchors []Anchor
}

// NewAnchorRegistry creates a new registry with default anchors
func NewAnchorRegistry() *AnchorRegistry {
	return &AnchorRegistry{
		anchors: []Anchor{
			NewAttachmentAnchor(),
			NewSMaskAnchor(),
			NewContentAnchor(),
		},
	}
}

// GetAvailableAnchors returns all registered anchors
func (r *AnchorRegistry) GetAvailableAnchors() []Anchor {
	return r.anchors
}

// GetAnchorByName returns an anchor by its name
func (r *AnchorRegistry) GetAnchorByName(name string) Anchor {
	for _, anchor := range r.anchors {
		if anchor.Name() == name {
			return anchor
		}
	}
	return nil
}

// AddAnchor registers a new anchor type
func (r *AnchorRegistry) AddAnchor(anchor Anchor) {
	r.anchors = append(r.anchors, anchor)
}
