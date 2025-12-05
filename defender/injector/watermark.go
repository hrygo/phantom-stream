package injector

import (
	"errors"
	"fmt"
	"os"
)

var (
	magicHeader = []byte{0xCA, 0xFE, 0xBA, 0xBE}
	keySize     = 32
	nonceSize   = 12
	attachName  = "font_license.txt"
)

var (
	// ErrInvalidKeySize indicates the encryption key is not the correct length
	ErrInvalidKeySize = fmt.Errorf("encryption key must be %d bytes long", keySize)
	// ErrInvalidPDFFile indicates the file is not a valid PDF
	ErrInvalidPDFFile = errors.New("invalid PDF file")
	// ErrShortPayload indicates the payload is too short to be valid
	ErrShortPayload = errors.New("payload too short")
	// ErrMagicHeaderMismatch indicates the magic header does not match
	ErrMagicHeaderMismatch = errors.New("magic header mismatch")
	// ErrAttachmentNotFound indicates the attachment was not found
	ErrAttachmentNotFound = errors.New("attachment not found")
)

// Sign embeds an encrypted message into a PDF file using triple-anchor strategy:
// Anchor 1 (Main): Attachment - Easy to detect but standard-compliant
// Anchor 2 (Stealth): Image SMask - Highly covert backup signature
// Anchor 3 (Phase 8): Content Stream Perturbation - Watermark bound to rendering
// Sign embeds an encrypted message into a PDF file using selected anchor strategies.
// selectedAnchors: list of anchor names to use. If empty, uses all available anchors.
func Sign(filePath, message, key, round string, selectedAnchors []string) error {
	// Validate inputs
	if err := validateInputs(filePath, message, key); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create crypto manager and encrypt payload
	crypto, err := NewCryptoManager([]byte(key))
	if err != nil {
		return fmt.Errorf("failed to create crypto manager: %w", err)
	}

	payload, err := crypto.Encrypt(message)
	if err != nil {
		return fmt.Errorf("failed to encrypt message: %w", err)
	}

	// Generate output file paths
	suffix := "_signed"
	if round != "" {
		suffix = "_" + round + "_signed"
	}
	tempOutputPath1, err := generateOutputPath(filePath, "_temp1")
	if err != nil {
		return fmt.Errorf("failed to generate temp1 output path: %w", err)
	}
	tempOutputPath2, err := generateOutputPath(filePath, "_temp2")
	if err != nil {
		return fmt.Errorf("failed to generate temp2 output path: %w", err)
	}
	finalOutputPath, err := generateOutputPath(filePath, suffix)
	if err != nil {
		return fmt.Errorf("failed to generate output path: %w", err)
	}

	// Get anchor registry
	registry := NewAnchorRegistry()
	allAnchors := registry.GetAvailableAnchors()

	// Filter anchors based on selection
	var anchorsToUse []Anchor
	if len(selectedAnchors) == 0 {
		anchorsToUse = allAnchors
	} else {
		for _, name := range selectedAnchors {
			for _, a := range allAnchors {
				if a.Name() == name {
					anchorsToUse = append(anchorsToUse, a)
					break
				}
			}
		}
	}

	if len(anchorsToUse) == 0 {
		return fmt.Errorf("no valid anchors selected")
	}

	anchorCount := 0
	anchorNames := []string{}
	currentInput := filePath
	// We need to swap between temp1 and temp2 for intermediate steps
	// Initial: Input -> Temp1
	// Step 2: Temp1 -> Temp2
	// Step 3: Temp2 -> Temp1
	// Final Step: -> FinalOutput

	// Helper to determine output for current step
	getOutput := func(step, total int) string {
		if step == total-1 {
			return finalOutputPath
		}
		if step%2 == 0 {
			return tempOutputPath1
		}
		return tempOutputPath2
	}

	for i, anchor := range anchorsToUse {
		output := getOutput(i, len(anchorsToUse))

		fmt.Printf("[*] Injecting Anchor %d/%d: %s...\n", i+1, len(anchorsToUse), anchor.Name())

		// Visual anchor displays plaintext; others use encrypted payload
		injectPayload := payload
		if anchor.Name() == "Visual" {
			injectPayload = []byte(message)
		}
		if err := anchor.Inject(currentInput, output, injectPayload); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ Warning: %s injection failed: %v\n", anchor.Name(), err)
			// If injection failed, we need to ensure the chain continues.
			// If this is the first step, we haven't created any temp file yet.
			// If intermediate, we have a temp file at currentInput.
			// We should copy currentInput to output to keep the chain, OR just skip this step's output
			// and use currentInput for the next step.

			// Strategy: Skip this anchor, don't update currentInput
			// But if this was the last anchor, we need to move currentInput to finalOutputPath
			if i == len(anchorsToUse)-1 {
				// If we have a previous temp file, rename it to final
				if currentInput != filePath {
					if err := os.Rename(currentInput, finalOutputPath); err != nil {
						return fmt.Errorf("failed to finalize output: %w", err)
					}
				} else {
					// No anchors succeeded? Or just first failed.
					// If first failed and it's the only one, we fail.
					return fmt.Errorf("failed to inject %s and it was the only anchor", anchor.Name())
				}
			}
			continue
		}

		// Injection successful
		if currentInput != filePath {
			os.Remove(currentInput) // Remove previous temp
		}
		currentInput = output
		anchorCount++
		anchorNames = append(anchorNames, anchor.Name())
		fmt.Printf("✓ Anchor %s embedded\n", anchor.Name())
	}

	// Clean up temp files (if any remain)
	os.Remove(tempOutputPath1)
	os.Remove(tempOutputPath2)

	if anchorCount == 0 {
		return fmt.Errorf("failed to inject any anchors")
	}

	// Report signature mode
	fmt.Printf("✓ Signature mode: %d-anchor strategy\n", anchorCount)
	for i, name := range anchorNames {
		fmt.Printf("  - Anchor %d: %s\n", i+1, name)
	}

	fmt.Printf("✓ Successfully signed PDF: %s\n", finalOutputPath)
	return nil
}

// Verify extracts and decrypts the hidden message from a signed PDF file.
// selectedAnchors: list of anchor names to verify. If empty, verifies all.
// Returns the extracted message and the name of the anchor that succeeded.
func Verify(filePath, key string, selectedAnchors []string) (string, string, error) {
	// Validate inputs
	if err := validateVerifyInputs(filePath, key); err != nil {
		return "", "", fmt.Errorf("validation failed: %w", err)
	}

	// Create crypto manager
	crypto, err := NewCryptoManager([]byte(key))
	if err != nil {
		return "", "", fmt.Errorf("failed to create crypto manager: %w", err)
	}

	// Get anchor registry
	registry := NewAnchorRegistry()
	allAnchors := registry.GetAvailableAnchors()

	// Filter anchors
	var anchorsToUse []Anchor
	if len(selectedAnchors) == 0 {
		anchorsToUse = allAnchors
	} else {
		for _, name := range selectedAnchors {
			for _, a := range allAnchors {
				if a.Name() == name {
					anchorsToUse = append(anchorsToUse, a)
					break
				}
			}
		}
	}

	if len(anchorsToUse) == 0 {
		return "", "", fmt.Errorf("no valid anchors selected")
	}

	// Try each anchor in order
	for _, anchor := range anchorsToUse {
		fmt.Fprintf(os.Stderr, "[DEBUG] Attempting Anchor: %s...\n", anchor.Name())

		payload, extractErr := anchor.Extract(filePath)
		if extractErr != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] %s: Extraction failed: %v\n", anchor.Name(), extractErr)
			continue
		}

		fmt.Fprintf(os.Stderr, "[DEBUG] %s: Extracted %d bytes\n", anchor.Name(), len(payload))

		// Decrypt and verify
		message, decryptErr := crypto.Decrypt(payload)
		if decryptErr == nil {
			fmt.Printf("✓ Verified via %s\n", anchor.Name())
			return message, anchor.Name(), nil
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] %s: Decryption failed: %v\n", anchor.Name(), decryptErr)
	}

	// All anchors failed
	return "", "", fmt.Errorf("verification failed: all selected anchors invalid or missing")
}

// Deprecated: Use CryptoManager.Encrypt instead
// Kept for backward compatibility with old tests
func createEncryptedPayload(message string, key []byte) ([]byte, error) {
	crypto, err := NewCryptoManager(key)
	if err != nil {
		return nil, err
	}
	return crypto.Encrypt(message)
}

// Deprecated: Use AttachmentAnchor.Extract instead
// Kept for backward compatibility with old tests
func ExtractPayloadFromPDF(filePath string) ([]byte, error) {
	anchor := NewAttachmentAnchor()
	return anchor.Extract(filePath)
}

// Deprecated: Use CryptoManager.Decrypt instead
// Kept for backward compatibility with old tests
func decryptPayload(payload, key []byte) (string, error) {
	crypto, err := NewCryptoManager(key)
	if err != nil {
		return "", err
	}
	return crypto.Decrypt(payload)
}

// Deprecated: Use SMaskAnchor.Inject instead
// Kept for backward compatibility with old tests
func InjectSMaskAnchor(inputPath, outputPath string, payload []byte) error {
	anchor := NewSMaskAnchor()
	return anchor.Inject(inputPath, outputPath, payload)
}

// Deprecated: Use SMaskAnchor.Extract instead
// Kept for backward compatibility with old tests
func ExtractSMaskPayloadFromPDF(filePath string) ([]byte, error) {
	anchor := NewSMaskAnchor()
	return anchor.Extract(filePath)
}
