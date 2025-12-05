package injector

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test constants
const (
	testKey32 = "12345678901234567890123456789012" // 32 bytes

	testMessage = "TestUser:Phase6-Optimized"
)

// TestValidateInputs tests the validateInputs function
func TestValidateInputs(t *testing.T) {
	// Create a temporary PDF file for testing
	tmpDir := t.TempDir()
	validPDF := filepath.Join(tmpDir, "test.pdf")
	if err := os.WriteFile(validPDF, []byte("%PDF-1.4\n"), 0600); err != nil {
		t.Fatalf("Failed to create test PDF: %v", err)
	}

	tests := []struct {
		name        string
		filePath    string
		message     string
		key         string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid inputs",
			filePath:    validPDF,
			message:     "test message",
			key:         testKey32,
			expectError: false,
		},
		{
			name:        "Empty file path",
			filePath:    "",
			message:     "test",
			key:         testKey32,
			expectError: true,
			errorMsg:    "file path cannot be empty",
		},
		{
			name:        "Empty message",
			filePath:    validPDF,
			message:     "",
			key:         testKey32,
			expectError: true,
			errorMsg:    "message cannot be empty",
		},
		{
			name:        "Invalid key size - too short",
			filePath:    validPDF,
			message:     "test",
			key:         "short",
			expectError: true,
			errorMsg:    "encryption key must be 32 bytes long",
		},
		{
			name:        "File does not exist",
			filePath:    "/nonexistent/file.pdf",
			message:     "test",
			key:         testKey32,
			expectError: true,
			errorMsg:    "file does not exist",
		},
		{
			name:        "Not a PDF file",
			filePath:    filepath.Join(tmpDir, "test.txt"),
			message:     "test",
			key:         testKey32,
			expectError: true,
			errorMsg:    "file does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateInputs(tt.filePath, tt.message, tt.key)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestCreateEncryptedPayload tests the createEncryptedPayload function
func TestCreateEncryptedPayload(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		key         []byte
		expectError bool
	}{
		{
			name:        "Valid encryption",
			message:     "test message",
			key:         []byte(testKey32),
			expectError: false,
		},
		{
			name:        "Empty message",
			message:     "",
			key:         []byte(testKey32),
			expectError: false,
		},
		{
			name:        "Long message",
			message:     strings.Repeat("A", 1000),
			key:         []byte(testKey32),
			expectError: false,
		},
		{
			name:        "Invalid key size",
			message:     "test",
			key:         []byte("short"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := createEncryptedPayload(tt.message, tt.key)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify payload structure
			minSize := len(magicHeader) + nonceSize
			if len(payload) < minSize {
				t.Errorf("Payload too short: got %d bytes, expected at least %d", len(payload), minSize)
			}

			// Verify magic header
			if !bytes.Equal(payload[:len(magicHeader)], magicHeader) {
				t.Error("Magic header mismatch")
			}

			// Verify we can decrypt it
			nonce := payload[len(magicHeader) : len(magicHeader)+nonceSize]
			encryptedMsg := payload[len(magicHeader)+nonceSize:]

			block, err := aes.NewCipher(tt.key)
			if err != nil {
				t.Fatalf("Failed to create cipher: %v", err)
			}

			gcm, err := cipher.NewGCM(block)
			if err != nil {
				t.Fatalf("Failed to create GCM: %v", err)
			}

			decrypted, err := gcm.Open(nil, nonce, encryptedMsg, nil)
			if err != nil {
				t.Fatalf("Failed to decrypt: %v", err)
			}

			if string(decrypted) != tt.message {
				t.Errorf("Decrypted message mismatch: got '%s', want '%s'", string(decrypted), tt.message)
			}
		})
	}
}

// TestDecryptPayload tests the decryptPayload function
func TestDecryptPayload(t *testing.T) {
	validKey := []byte(testKey32)
	wrongKey := []byte("wrongkey1234567890123456789012")

	// Create a valid payload
	validPayload, err := createEncryptedPayload(testMessage, validKey)
	if err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
	}

	tests := []struct {
		name        string
		payload     []byte
		key         []byte
		expectError bool
		errorType   error
	}{
		{
			name:        "Valid decryption",
			payload:     validPayload,
			key:         validKey,
			expectError: false,
		},
		{
			name:        "Wrong key",
			payload:     validPayload,
			key:         wrongKey,
			expectError: true,
		},
		{
			name:        "Payload too short",
			payload:     []byte{0xCA, 0xFE},
			key:         validKey,
			expectError: true,
			errorType:   ErrShortPayload,
		},
		{
			name:        "Invalid magic header",
			payload:     append([]byte{0xFF, 0xFF, 0xFF, 0xFF}, validPayload[4:]...),
			key:         validKey,
			expectError: true,
			errorType:   ErrMagicHeaderMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, err := decryptPayload(tt.payload, tt.key)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errorType != nil && err != tt.errorType {
					if !strings.Contains(err.Error(), tt.errorType.Error()) {
						t.Errorf("Expected error type '%v', got '%v'", tt.errorType, err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if message != testMessage {
				t.Errorf("Message mismatch: got '%s', want '%s'", message, testMessage)
			}
		})
	}
}

// TestGenerateOutputPath tests the generateOutputPath function
func TestGenerateOutputPath(t *testing.T) {
	tests := []struct {
		name        string
		inputPath   string
		suffix      string
		expectError bool
		expectedExt string
	}{
		{
			name:        "Simple PDF path",
			inputPath:   "/path/to/file.pdf",
			suffix:      "_signed",
			expectError: false,
			expectedExt: ".pdf",
		},
		{
			name:        "Path with multiple dots",
			inputPath:   "/path/to/file.name.pdf",
			suffix:      "_signed",
			expectError: false,
			expectedExt: ".pdf",
		},
		{
			name:        "No directory path",
			inputPath:   "file.pdf",
			suffix:      "_signed",
			expectError: false,
			expectedExt: ".pdf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := generateOutputPath(tt.inputPath, tt.suffix)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !strings.HasSuffix(output, tt.expectedExt) {
				t.Errorf("Output path doesn't have expected extension: got '%s', want suffix '%s'", output, tt.expectedExt)
			}

			if !strings.Contains(output, tt.suffix) {
				t.Errorf("Output path doesn't contain suffix '%s': got '%s'", tt.suffix, output)
			}

			inputDir := filepath.Dir(tt.inputPath)
			outputDir := filepath.Dir(output)
			if inputDir != outputDir {
				t.Errorf("Output directory mismatch: got '%s', want '%s'", outputDir, inputDir)
			}
		})
	}
}

// TestEncryptionDecryptionRoundTrip tests full encryption/decryption cycle
func TestEncryptionDecryptionRoundTrip(t *testing.T) {
	testCases := []string{
		"Simple message",
		"UserID:12345",
		"",
		"Message with unicode: 中文",
	}

	key := []byte(testKey32)

	for _, msg := range testCases {
		t.Run("Message length: "+string(rune(len(msg))), func(t *testing.T) {
			payload, err := createEncryptedPayload(msg, key)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			decrypted, err := decryptPayload(payload, key)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			if decrypted != msg {
				t.Errorf("Round-trip failed: got '%s', want '%s'", decrypted, msg)
			}
		})
	}
}

// BenchmarkCreateEncryptedPayload benchmarks encryption
func BenchmarkCreateEncryptedPayload(b *testing.B) {
	key := []byte(testKey32)
	message := "BenchmarkMessage:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := createEncryptedPayload(message, key)
		if err != nil {
			b.Fatalf("Encryption failed: %v", err)
		}
	}
}

// BenchmarkDecryptPayload benchmarks decryption
func BenchmarkDecryptPayload(b *testing.B) {
	key := []byte(testKey32)
	message := "BenchmarkMessage:12345"

	payload, err := createEncryptedPayload(message, key)
	if err != nil {
		b.Fatalf("Failed to create test payload: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := decryptPayload(payload, key)
		if err != nil {
			b.Fatalf("Decryption failed: %v", err)
		}
	}
}
