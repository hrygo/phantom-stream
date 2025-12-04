package injector

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
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

// Sign embeds an encrypted message into a PDF file as an attachment.
// It returns the output file path on success.
func Sign(filePath, message, key, round string) error {
	// Validate inputs
	if err := validateInputs(filePath, message, key); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Create encrypted payload
	payload, err := createEncryptedPayload(message, []byte(key))
	if err != nil {
		return fmt.Errorf("failed to create encrypted payload: %w", err)
	}

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

	// Generate output file path
	suffix := "_signed"
	if round != "" {
		suffix = "_" + round + "_signed"
	}
	outputFilePath, err := generateOutputPath(filePath, suffix)
	if err != nil {
		return fmt.Errorf("failed to generate output path: %w", err)
	}

	// Add attachment to PDF
	conf := model.NewDefaultConfiguration()
	if err := api.AddAttachmentsFile(filePath, outputFilePath, []string{payloadPath}, true, conf); err != nil {
		return fmt.Errorf("failed to add attachment to PDF: %w", err)
	}

	fmt.Printf("✓ Successfully signed PDF: %s\n", outputFilePath)
	fmt.Printf("  - Attachment: %s (size: %d bytes)\n", attachName, len(payload))
	return nil
}

// Verify extracts and decrypts the hidden message from a signed PDF file.
func Verify(filePath, key string) (string, error) {
	// Validate inputs
	if err := validateVerifyInputs(filePath, key); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// Extract payload from PDF
	payload, err := extractPayloadFromPDF(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to extract payload: %w", err)
	}

	// Decrypt and verify payload
	message, err := decryptPayload(payload, []byte(key))
	if err != nil {
		return "", fmt.Errorf("failed to decrypt payload: %w", err)
	}

	return message, nil
}

// Helper functions

// validateInputs validates the inputs for the Sign function
func validateInputs(filePath, message, key string) error {
	if filePath == "" {
		return errors.New("file path cannot be empty")
	}
	if message == "" {
		return errors.New("message cannot be empty")
	}
	if len(key) != keySize {
		return ErrInvalidKeySize
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	// Check if it's a PDF file
	if !strings.HasSuffix(strings.ToLower(filePath), ".pdf") {
		return ErrInvalidPDFFile
	}

	return nil
}

// validateVerifyInputs validates the inputs for the Verify function
func validateVerifyInputs(filePath, key string) error {
	if filePath == "" {
		return errors.New("file path cannot be empty")
	}
	if len(key) != keySize {
		return ErrInvalidKeySize
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", filePath)
	}

	return nil
}

// createEncryptedPayload creates an encrypted payload from the message
func createEncryptedPayload(message string, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	encryptedMessage := gcm.Seal(nil, nonce, []byte(message), nil)

	// Build payload: magic header + nonce + encrypted message
	payload := make([]byte, 0, len(magicHeader)+len(nonce)+len(encryptedMessage))
	payload = append(payload, magicHeader...)
	payload = append(payload, nonce...)
	payload = append(payload, encryptedMessage...)

	return payload, nil
}

// extractPayloadFromPDF extracts the payload attachment from a PDF file
func extractPayloadFromPDF(filePath string) ([]byte, error) {
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

// decryptPayload decrypts the payload and returns the message
func decryptPayload(payload, key []byte) (string, error) {
	// Validate payload structure
	minSize := len(magicHeader) + nonceSize
	if len(payload) < minSize {
		return "", ErrShortPayload
	}

	// Verify magic header
	for i := range magicHeader {
		if payload[i] != magicHeader[i] {
			return "", ErrMagicHeaderMismatch
		}
	}

	// Extract nonce and encrypted message
	nonce := payload[len(magicHeader) : len(magicHeader)+nonceSize]
	encryptedMessage := payload[len(magicHeader)+nonceSize:]

	// Decrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	decrypted, err := gcm.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed (wrong key or corrupted data): %w", err)
	}

	return string(decrypted), nil
}

// generateOutputPath generates the output file path with a suffix
func generateOutputPath(inputPath, suffix string) (string, error) {
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	if name == "" {
		return "", errors.New("invalid file name")
	}

	outputFileName := name + suffix + ext
	return filepath.Join(dir, outputFileName), nil
}
