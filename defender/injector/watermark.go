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

// Sign embeds an encrypted message into a PDF file using dual-anchor strategy:
// Anchor 1 (Main): Attachment - Easy to detect but standard-compliant
// Anchor 2 (Stealth): Image SMask - Highly covert backup signature
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
	tempOutputPath, err := generateOutputPath(filePath, "_temp")
	if err != nil {
		return fmt.Errorf("failed to generate temp output path: %w", err)
	}
	finalOutputPath, err := generateOutputPath(filePath, suffix)
	if err != nil {
		return fmt.Errorf("failed to generate output path: %w", err)
	}

	// Get anchor registry
	registry := NewAnchorRegistry()
	anchors := registry.GetAvailableAnchors()

	// === Anchor 1: Attachment (Main Anchor) ===
	attachmentAnchor := anchors[0] // AttachmentAnchor
	if err := attachmentAnchor.Inject(filePath, tempOutputPath, payload); err != nil {
		return fmt.Errorf("failed to inject attachment anchor: %w", err)
	}
	fmt.Printf("✓ Anchor 1/2: %s embedded (%d bytes)\n", attachmentAnchor.Name(), len(payload))

	// === Anchor 2: SMask (Stealth Anchor) ===
	smaskAnchor := anchors[1] // SMaskAnchor
	if err := smaskAnchor.Inject(tempOutputPath, finalOutputPath, payload); err != nil {
		// SMask injection failed - fallback to attachment-only mode
		fmt.Fprintf(os.Stderr, "⚠ Warning: SMask injection failed, using attachment-only mode: %v\n", err)
		if renameErr := os.Rename(tempOutputPath, finalOutputPath); renameErr != nil {
			return fmt.Errorf("failed to finalize output: %w", renameErr)
		}
		fmt.Printf("✓ Signature mode: Single-anchor (Attachment only)\n")
	} else {
		os.Remove(tempOutputPath)
		fmt.Printf("✓ Anchor 2/2: %s embedded\n", smaskAnchor.Name())
		fmt.Printf("✓ Signature mode: Dual-anchor (Attachment + SMask)\n")
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
