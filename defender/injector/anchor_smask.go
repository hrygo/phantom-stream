package injector

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// SMaskAnchor implements signature embedding via image SMask (soft mask)
type SMaskAnchor struct{}

// NewSMaskAnchor creates a new SMask anchor
func NewSMaskAnchor() *SMaskAnchor {
	return &SMaskAnchor{}
}

// Name returns the anchor type name
func (a *SMaskAnchor) Name() string {
	return "SMask"
}

// Inject embeds the payload into a PDF via image SMask
func (a *SMaskAnchor) Inject(inputPath, outputPath string, payload []byte) error {
	// Read and parse PDF
	ctx, err := api.ReadContextFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read PDF context: %w", err)
	}

	// Create SMask injector
	injector := &smaskInjector{payload: payload}

	// Inject SMask
	if err := injector.inject(ctx); err != nil {
		return err
	}

	// Write modified PDF
	if err := api.WriteContextFile(ctx, outputPath); err != nil {
		return fmt.Errorf("failed to write PDF: %w", err)
	}

	return nil
}

// Extract retrieves the payload from SMask anchor
func (a *SMaskAnchor) Extract(filePath string) ([]byte, error) {
	// Read and parse PDF
	ctx, err := api.ReadContextFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF context: %w", err)
	}

	// Extract SMask payload
	extractor := &smaskExtractor{}
	payload, err := extractor.extract(ctx)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

// IsAvailable checks if SMask anchor can be used
// Requires at least one image in the PDF
func (a *SMaskAnchor) IsAvailable(ctx *model.Context) bool {
	images := findImageXObjects(ctx)
	return len(images) > 0
}

// smaskInjector handles SMask injection logic
type smaskInjector struct {
	payload []byte
}

// inject injects the payload into a PDF via image SMask
func (s *smaskInjector) inject(ctx *model.Context) error {
	// Find all image XObjects in the PDF
	images := findImageXObjects(ctx)

	if len(images) == 0 {
		return fmt.Errorf("no images found in PDF (SMask anchor requires at least one image)")
	}

	// Use the first image as anchor
	targetImgRef := images[0]
	targetImg, err := getImageObject(ctx, targetImgRef)
	if err != nil {
		return fmt.Errorf("failed to get image object: %w", err)
	}

	// Get image dimensions
	width, height, err := getImageDimensions(&targetImg)
	if err != nil {
		return fmt.Errorf("failed to get image dimensions: %w", err)
	}

	// Create SMask object
	smaskRef, err := s.createSMaskObject(ctx, width, height)
	if err != nil {
		return fmt.Errorf("failed to create SMask object: %w", err)
	}

	// Update the image object in xRefTable
	entry, found := ctx.Find(int(targetImgRef.ObjectNumber))
	if !found {
		return fmt.Errorf("image object not found in xRefTable")
	}

	if entry.Object == nil {
		return fmt.Errorf("image object is nil")
	}

	// Handle pointer vs value type
	if sdPtr, ok := entry.Object.(*types.StreamDict); ok {
		sdPtr.Update("SMask", *smaskRef)
	} else {
		// entry.Object is a value, must recreate
		actualImg, ok := entry.Object.(types.StreamDict)
		if !ok {
			return fmt.Errorf("object is not a StreamDict, got: %T", entry.Object)
		}

		// Create a new Dict with all existing entries PLUS the SMask
		newDict := types.NewDict()
		for k, v := range actualImg.Dict {
			newDict[k] = v
		}
		newDict["SMask"] = *smaskRef

		// Create a new StreamDict with the new Dict
		newStreamDict := types.StreamDict{
			Dict:              newDict,
			StreamOffset:      actualImg.StreamOffset,
			StreamLength:      actualImg.StreamLength,
			StreamLengthObjNr: actualImg.StreamLengthObjNr,
			FilterPipeline:    actualImg.FilterPipeline,
			Raw:               actualImg.Raw,
			Content:           actualImg.Content,
		}

		entry.Object = newStreamDict
	}

	return nil
}

// createSMaskObject creates a new SMask stream object with embedded payload
func (s *smaskInjector) createSMaskObject(ctx *model.Context, width, height int) (*types.IndirectRef, error) {
	// Create mask data: all 255 (fully opaque)
	maskSize := width * height
	maskData := make([]byte, maskSize)
	for i := range maskData {
		maskData[i] = 255
	}

	// Prepare payload with magic header
	magicHeaderCopy := make([]byte, len(magicHeader))
	copy(magicHeaderCopy, magicHeader)
	fullPayload := append(magicHeaderCopy, s.payload...)

	// Embed payload at the end of mask data
	payloadOffset := len(maskData) - len(fullPayload)
	if payloadOffset < 100 {
		return nil, fmt.Errorf("image too small for payload (need at least %d bytes, have %d)",
			len(fullPayload)+100, maskSize)
	}

	copy(maskData[payloadOffset:], fullPayload)

	// Compress mask data with Flate (zlib)
	compressedData, err := compressFlate(maskData)
	if err != nil {
		return nil, fmt.Errorf("failed to compress mask data: %w", err)
	}

	// Create SMask stream dictionary
	streamLength := int64(len(compressedData))
	smaskDict := types.NewStreamDict(
		types.NewDict(),
		0,
		&streamLength,
		nil,
		[]types.PDFFilter{},
	)

	// Set raw content (compressed)
	smaskDict.Raw = compressedData
	smaskDict.Content = maskData

	smaskDict.InsertName("Type", "XObject")
	smaskDict.InsertName("Subtype", "Image")
	smaskDict.InsertInt("Width", width)
	smaskDict.InsertInt("Height", height)
	smaskDict.InsertName("ColorSpace", "DeviceGray")
	smaskDict.InsertInt("BitsPerComponent", 8)
	smaskDict.InsertName("Filter", "FlateDecode")

	// Add SMask to xRefTable
	objNr, err := ctx.InsertObject(smaskDict)
	if err != nil {
		return nil, fmt.Errorf("failed to insert SMask object: %w", err)
	}

	ref := types.NewIndirectRef(objNr, 0)
	return ref, nil
}

// smaskExtractor handles SMask extraction logic
type smaskExtractor struct{}

// extract extracts payload from SMask anchor
func (e *smaskExtractor) extract(ctx *model.Context) ([]byte, error) {
	// Find all image XObjects
	images := findImageXObjects(ctx)

	fmt.Fprintf(os.Stderr, "[DEBUG] SMask: Found %d images\n", len(images))

	// Search for SMask in images
	for idx, imgRef := range images {
		fmt.Fprintf(os.Stderr, "[DEBUG] SMask: Checking image %d (object %d)\n", idx+1, imgRef.ObjectNumber)
		obj, err := ctx.Dereference(imgRef)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] SMask: Dereference failed: %v\n", err)
			continue
		}

		streamDict, ok := obj.(types.StreamDict)
		if !ok {
			fmt.Fprintf(os.Stderr, "[DEBUG] SMask: Not a StreamDict\n")
			continue
		}

		// Check if image has SMask
		smaskRef := streamDict.IndirectRefEntry("SMask")
		if smaskRef == nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] SMask: No SMask ref\n")
			continue
		}

		fmt.Fprintf(os.Stderr, "[DEBUG] SMask: Found SMask ref -> object %d\n", smaskRef.ObjectNumber)

		// Get SMask object
		smaskObj, err := ctx.Dereference(*smaskRef)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] SMask: Dereference SMask failed: %v\n", err)
			continue
		}

		smaskStream, ok := smaskObj.(types.StreamDict)
		if !ok {
			fmt.Fprintf(os.Stderr, "[DEBUG] SMask: SMask not a StreamDict\n")
			continue
		}

		// Decode SMask stream
		maskData, err := decodeSMask(&smaskStream)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] SMask: Decode failed: %v\n", err)
			continue
		}

		fmt.Fprintf(os.Stderr, "[DEBUG] SMask: Decoded %d bytes\n", len(maskData))

		// Extract payload from end of mask data
		payload, err := e.findPayloadInMaskData(maskData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] SMask: %v\n", err)
			continue
		}

		return payload, nil
	}

	return nil, fmt.Errorf("SMask payload not found")
}

// findPayloadInMaskData scans mask data for payload (magic header)
func (e *smaskExtractor) findPayloadInMaskData(maskData []byte) ([]byte, error) {
	// Scan backwards for magic header
	maxScanSize := 500
	if len(maskData) < maxScanSize {
		maxScanSize = len(maskData)
	}

	scanStart := len(maskData) - maxScanSize
	scanData := maskData[scanStart:]

	// Find magic header
	for i := 0; i <= len(scanData)-len(magicHeader); i++ {
		if bytes.Equal(scanData[i:i+len(magicHeader)], magicHeader) {
			payloadStart := scanStart + i + len(magicHeader) // Skip magic header
			payload := maskData[payloadStart:]
			fmt.Fprintf(os.Stderr, "[DEBUG] SMask: Found magic header at offset %d, payload size %d\n",
				payloadStart-len(magicHeader), len(payload))
			return payload, nil
		}
	}

	return nil, fmt.Errorf("magic header not found in last %d bytes", maxScanSize)
}

// PDF utility functions (shared by injector and extractor)

// findImageXObjects finds all image XObjects in the PDF
func findImageXObjects(ctx *model.Context) []types.IndirectRef {
	var images []types.IndirectRef

	for objNr := 1; objNr <= *ctx.XRefTable.Size; objNr++ {
		entry, found := ctx.Find(objNr)
		if !found || entry.Free || entry.Compressed {
			continue
		}

		obj, err := ctx.Dereference(types.IndirectRef{ObjectNumber: types.Integer(objNr), GenerationNumber: types.Integer(0)})
		if err != nil {
			continue
		}

		if streamDict, ok := obj.(types.StreamDict); ok {
			if streamDict.Type() != nil && *streamDict.Type() == "XObject" {
				subtype := streamDict.NameEntry("Subtype")
				if subtype != nil && *subtype == "Image" {
					images = append(images, types.IndirectRef{ObjectNumber: types.Integer(objNr), GenerationNumber: types.Integer(0)})
				}
			}
		}
	}

	return images
}

// getImageObject retrieves the image stream dictionary
func getImageObject(ctx *model.Context, ref types.IndirectRef) (types.StreamDict, error) {
	obj, err := ctx.Dereference(ref)
	if err != nil {
		return types.StreamDict{}, err
	}

	streamDict, ok := obj.(types.StreamDict)
	if !ok {
		return types.StreamDict{}, fmt.Errorf("object is not a stream dictionary")
	}

	return streamDict, nil
}

// getImageDimensions extracts width and height from image dictionary
func getImageDimensions(img *types.StreamDict) (width, height int, err error) {
	widthObj := img.IntEntry("Width")
	heightObj := img.IntEntry("Height")

	if widthObj == nil || heightObj == nil {
		return 0, 0, fmt.Errorf("image dimensions not found")
	}

	return *widthObj, *heightObj, nil
}

// compressFlate compresses data using zlib (Flate)
func compressFlate(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)

	if _, err := w.Write(data); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decodeSMask decodes SMask stream data (internal helper)
func decodeSMask(stream *types.StreamDict) ([]byte, error) {
	rawData := stream.Raw

	// Check if compressed with Flate
	filterName := stream.NameEntry("Filter")
	if filterName != nil && *filterName == "FlateDecode" {
		// Decompress
		reader, err := zlib.NewReader(bytes.NewReader(rawData))
		if err != nil {
			return nil, fmt.Errorf("failed to create zlib reader: %w", err)
		}
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress: %w", err)
		}

		return decompressed, nil
	}

	// Not compressed
	return rawData, nil
}
