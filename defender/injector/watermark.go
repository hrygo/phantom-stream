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
func Sign(filePath, message, key, round string) error {
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
	anchors := registry.GetAvailableAnchors()

	anchorCount := 0
	anchorNames := []string{}
	currentInput := filePath
	currentOutput := tempOutputPath1

	// === Anchor 1: Attachment (Main Anchor) ===
	attachmentAnchor := anchors[0] // AttachmentAnchor
	if err := attachmentAnchor.Inject(currentInput, currentOutput, payload); err != nil {
		return fmt.Errorf("failed to inject attachment anchor: %w", err)
	}
	anchorCount++
	anchorNames = append(anchorNames, attachmentAnchor.Name())
	fmt.Printf("✓ Anchor %d/3: %s embedded (%d bytes)\n", anchorCount, attachmentAnchor.Name(), len(payload))
	currentInput = currentOutput
	currentOutput = tempOutputPath2

	// === Anchor 2: SMask (Stealth Anchor) ===
	smaskAnchor := anchors[1] // SMaskAnchor
	if err := smaskAnchor.Inject(currentInput, currentOutput, payload); err != nil {
		// SMask injection failed - continue without it
		fmt.Fprintf(os.Stderr, "⚠ Warning: SMask injection failed: %v\n", err)
		// Keep currentInput unchanged, use tempOutputPath2 for next anchor
		currentOutput = tempOutputPath2
	} else {
		os.Remove(currentInput)
		anchorCount++
		anchorNames = append(anchorNames, smaskAnchor.Name())
		fmt.Printf("✓ Anchor %d/3: %s embedded\n", anchorCount, smaskAnchor.Name())
		currentInput = currentOutput
		currentOutput = finalOutputPath
	}

	// === Anchor 3: Content Stream Perturbation (Phase 8) ===
	if len(anchors) > 2 {
		contentAnchor := anchors[2] // ContentAnchor
		if err := contentAnchor.Inject(currentInput, finalOutputPath, payload); err != nil {
			// Content injection failed - finalize with current anchors
			fmt.Fprintf(os.Stderr, "⚠ Warning: Content stream injection failed: %v\n", err)
			if err := os.Rename(currentInput, finalOutputPath); err != nil {
				return fmt.Errorf("failed to finalize output: %w", err)
			}
		} else {
			os.Remove(currentInput)
			anchorCount++
			anchorNames = append(anchorNames, contentAnchor.Name())
			fmt.Printf("✓ Anchor %d/3: %s embedded\n", anchorCount, contentAnchor.Name())
		}
	} else {
		// No content anchor available, finalize
		if err := os.Rename(currentInput, finalOutputPath); err != nil {
			return fmt.Errorf("failed to finalize output: %w", err)
		}
	}

	// Clean up temp files
	os.Remove(tempOutputPath1)
	os.Remove(tempOutputPath2)

	// Report signature mode
	switch anchorCount {
	case 1:
		fmt.Printf("✓ Signature mode: Single-anchor (%s only)\n", anchorNames[0])
	case 2:
		fmt.Printf("✓ Signature mode: Dual-anchor (%s + %s)\n", anchorNames[0], anchorNames[1])
	case 3:
		fmt.Printf("✓ Signature mode: Triple-anchor (%s + %s + %s) [Phase 8]\n", anchorNames[0], anchorNames[1], anchorNames[2])
	}

	fmt.Printf("✓ Successfully signed PDF: %s\n", finalOutputPath)
	return nil
}

// Verify extracts and decrypts the hidden message from a signed PDF file.
// Uses multi-anchor verification: checks both attachment and SMask anchors.
// Verification succeeds if ANY anchor is valid (fault-tolerant design).
func Verify(filePath, key string) (string, error) {
	// Validate inputs
	if err := validateVerifyInputs(filePath, key); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// Create crypto manager
	crypto, err := NewCryptoManager([]byte(key))
	if err != nil {
		return "", fmt.Errorf("failed to create crypto manager: %w", err)
	}

	// Get anchor registry
	registry := NewAnchorRegistry()
	anchors := registry.GetAvailableAnchors()

	// Try each anchor in order
	for idx, anchor := range anchors {
		fmt.Fprintf(os.Stderr, "[DEBUG] Attempting Anchor %d: %s...\n", idx+1, anchor.Name())

		payload, extractErr := anchor.Extract(filePath)
		if extractErr != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] Anchor %d: Extraction failed: %v\n", idx+1, extractErr)
			continue
		}

		fmt.Fprintf(os.Stderr, "[DEBUG] Anchor %d: Extracted %d bytes\n", idx+1, len(payload))

		// Decrypt and verify
		message, decryptErr := crypto.Decrypt(payload)
		if decryptErr == nil {
			if idx == 0 {
				fmt.Printf("✓ Verified via Anchor %d: %s\n", idx+1, anchor.Name())
			} else {
				fmt.Printf("✓ Verified via Anchor %d: %s (backup anchor activated)\n", idx+1, anchor.Name())
			}
			return message, nil
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] Anchor %d: Decryption failed: %v\n", idx+1, decryptErr)
	}

	// All anchors failed
	return "", fmt.Errorf("verification failed: all anchors invalid or missing")
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
func extractPayloadFromPDF(filePath string) ([]byte, error) {
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
func injectSMaskAnchor(inputPath, outputPath string, payload []byte) error {
	anchor := NewSMaskAnchor()
	return anchor.Inject(inputPath, outputPath, payload)
}

// Deprecated: Use SMaskAnchor.Extract instead
// Kept for backward compatibility with old tests
func extractSMaskPayloadFromPDF(filePath string) ([]byte, error) {
	anchor := NewSMaskAnchor()
	return anchor.Extract(filePath)
}
