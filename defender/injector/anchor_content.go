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
	contentMagicHeaderCopy := make([]byte, len(contentMagicHeader))
	copy(contentMagicHeaderCopy, contentMagicHeader)
	fullPayload := append(contentMagicHeaderCopy, payload...)
	fmt.Printf("[DEBUG] Content: Payload size %d bytes\n", len(fullPayload))

	injectedCount := 0

	// Iterate through all pages
	for i := 1; i <= ctx.PageCount; i++ {
		// Get page dictionary
		pageDict, _, _, err := ctx.PageDict(i, false)
		if err != nil {
			fmt.Printf("[DEBUG] Content: Failed to get page dict for page %d: %v\n", i, err)
			continue
		}

		// 1. Ensure a standard font (Helvetica) is available in Resources
		fontName := "/PhantomHelv" // Unique name to avoid conflict

		// Dereference Resources dict
		var resDict types.Dict
		if resObj, ok := pageDict["Resources"]; ok {
			resDict, err = ctx.XRefTable.DereferenceDict(resObj)
			if err != nil {
				fmt.Printf("[DEBUG] Content: Failed to dereference Resources for page %d: %v\n", i, err)
				continue
			}
		} else {
			// Create new Resources if missing
			resDict = types.NewDict()
			pageDict["Resources"] = resDict
		}

		// Ensure Font dict exists
		var fontDict types.Dict
		if fontObj, ok := resDict["Font"]; ok {
			fontDict, err = ctx.XRefTable.DereferenceDict(fontObj)
			if err != nil {
				fmt.Printf("[DEBUG] Content: Failed to dereference Font dict for page %d: %v\n", i, err)
				continue
			}
		} else {
			fontDict = types.NewDict()
			resDict["Font"] = fontDict
		}

		// We need to create a Type1 Font object for Helvetica
		fontObj := types.NewDict()
		fontObj.InsertName("Type", "Font")
		fontObj.InsertName("Subtype", "Type1")
		fontObj.InsertName("BaseFont", "Helvetica")

		// Add font object to XRefTable
		fontIndRef, err := ctx.XRefTable.IndRefForNewObject(fontObj)
		if err != nil {
			fmt.Printf("[DEBUG] Content: Failed to create font object: %v\n", err)
			continue
		}

		// Register in page resources
		fontDict[string(types.Name(fontName[1:]))] = *fontIndRef // Remove leading slash for key

		// 2. Create a NEW content stream with our payload
		var sb strings.Builder
		// Save graphics state (q), Begin Text (BT), Set Font (Tf), Invisible Mode (3 Tr)
		sb.WriteString(fmt.Sprintf("q\nBT\n%s 1 Tf\n3 Tr\n[", fontName))
		for _, b := range fullPayload {
			sb.WriteString(fmt.Sprintf(" ( ) %d", b))
		}
		sb.WriteString(" ] TJ\nET\nQ\n")

		contentData := []byte(sb.String())

		// Create stream dict
		sd := types.NewStreamDict(types.NewDict(), 0, nil, nil, nil)
		sd.Content = contentData

		// Compress
		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)
		if _, writeErr := w.Write(contentData); writeErr != nil {
			w.Close()
			fmt.Printf("[DEBUG] Content: Failed to write to zlib writer: %v\n", writeErr)
			continue
		}
		w.Close()

		sd.Raw = buf.Bytes()
		sd.InsertName("Filter", "FlateDecode")

		// Set correct compressed length
		compressedLen := int64(len(sd.Raw))
		sd.Insert("Length", types.Integer(compressedLen))
		sd.StreamLength = &compressedLen

		// Add stream to XRefTable
		streamIndRef, err := ctx.XRefTable.IndRefForNewObject(sd)
		if err != nil {
			fmt.Printf("[DEBUG] Content: Failed to create stream object: %v\n", err)
			continue
		}

		// 3. Append new stream to page Contents
		if contentObj, ok := pageDict["Contents"]; ok {
			switch obj := contentObj.(type) {
			case types.IndirectRef:
				// Convert single ref to array: [OldRef, NewRef]
				arr := types.Array{obj, *streamIndRef}
				pageDict["Contents"] = arr
			case types.Array:
				// Append to existing array
				obj = append(obj, *streamIndRef)
				pageDict["Contents"] = obj
			default:
				// Unknown type, overwrite (risky) or skip
				fmt.Printf("[DEBUG] Content: Unknown Contents type for page %d\n", i)
				continue
			}
		} else {
			// No contents, set as single ref
			pageDict["Contents"] = *streamIndRef
		}

		fmt.Printf("[DEBUG] Content: Injected new stream into page %d using font %s\n", i, fontName)
		injectedCount++
	}

	if injectedCount == 0 {
		return fmt.Errorf("failed to inject into any page")
	}

	fmt.Printf("[DEBUG] Content: Injected into %d pages\n", injectedCount)

	// Write output
	fmt.Printf("[DEBUG] Content: Writing output to %s\n", outputPath)
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
