package injector

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"bytes"
)

var (
	magicHeader = []byte{0xCA, 0xFE, 0xBA, 0xBE} // Magic Header for identifying hidden data
	// eofMarker   = []byte("%%EOF")                      // PDF End Of File marker
	keySize     = 32                               // AES-256 key size in bytes
	nonceSize   = 12                               // GCM standard nonce size
)

// Sign embeds an encrypted message into a PDF file.
// It takes the path to the source PDF, the message to embed, and an encryption key.
// A new PDF file named "{original_filename}_signed.pdf" will be created.
func Sign(filePath, message, key string) error {
	// 1. Key validation
	if len(key) != keySize {
		return fmt.Errorf("encryption key must be %d bytes long", keySize)
	}
	encryptionKey := []byte(key)

	// 2. Read original PDF content
	originalPDF, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read original PDF file: %w", err)
	}

	// 3. Create AES-256-GCM cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// 4. Generate random Nonce
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 5. Encrypt the message
	// The GCM seal function appends the authentication tag to the ciphertext.
	encryptedMessage := gcm.Seal(nil, nonce, []byte(message), nil)

	// 6. Build Payload: Magic Header + Nonce + Encrypted Message
	payload := make([]byte, 0, len(magicHeader)+len(nonce)+len(encryptedMessage))
	payload = append(payload, magicHeader...)
	payload = append(payload, nonce...)
	payload = append(payload, encryptedMessage...)

	// 7. Combine original PDF with newline and Payload
	// Add a newline to ensure %%EOF is on its own line before the payload
	finalContent := make([]byte, 0, len(originalPDF)+1+len(payload))
	finalContent = append(finalContent, originalPDF...)
	finalContent = append(finalContent, '\n') // Add newline as per PRD
	finalContent = append(finalContent, payload...)

	// 8. Write to new file
	outputFileName := fmt.Sprintf("%s_signed.pdf", filepath.Base(filePath)[:len(filepath.Base(filePath))-len(filepath.Ext(filePath))])
	outputFilePath := filepath.Join(filepath.Dir(filePath), outputFileName)

	if err := ioutil.WriteFile(outputFilePath, finalContent, 0644); err != nil {
		return fmt.Errorf("failed to write signed PDF file: %w", err)
	}

	return nil
}

// Verify extracts and verifies a hidden message from a PDF file.
// It returns the extracted message or an error if verification fails.
func Verify(filePath, key string) (string, error) {
	// 1. Key validation
	if len(key) != keySize {
		return "", fmt.Errorf("decryption key must be %d bytes long", keySize)
	}
	decryptionKey := []byte(key)

	// 2. Read PDF file content
	signedPDF, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read signed PDF file: %w", err)
	}

	// 3. Find the last %%EOF marker
	eofIndex := bytes.LastIndex(signedPDF, []byte("%%EOF"))
	if eofIndex == -1 {
		return "", errors.New("no %%EOF marker found in the file, likely not a valid PDF or tampered")
	}

	// Find the start of the payload after the last %%EOF marker.
	// The payload starts with magicHeader. We need to locate it after %%EOF.
	payloadStart := -1
	searchStart := eofIndex + len("%%EOF")

	// Iterate from searchStart to find the magicHeader
	for i := searchStart; i <= len(signedPDF)-len(magicHeader); i++ {
		if bytes.Equal(signedPDF[i:i+len(magicHeader)], magicHeader) {
			payloadStart = i
			break
		}
	}

	if payloadStart == -1 {
		return "", errors.New("magic header not found after %%EOF marker")
	}

	potentialPayload := signedPDF[payloadStart:]

	// 4. Validate payload length and Magic Header
	fmt.Printf("DEBUG: Potential payload starts with: %x\n", potentialPayload[:min(len(potentialPayload), 16)])
	if len(potentialPayload) < len(magicHeader)+nonceSize {
		return "", errors.New("payload too short to contain magic header and nonce")
	}

	// The magic header check is now implicitly part of finding payloadStart, but we keep it for redundancy and clarity.
	if !bytes.Equal(potentialPayload[0:len(magicHeader)], magicHeader) {
		return "", errors.New("magic header mismatch, should not happen if payloadStart is correct")
	}

	// 5. Extract Nonce and Ciphertext
	nonce := potentialPayload[len(magicHeader) : len(magicHeader)+nonceSize]
	encryptedMessage := potentialPayload[len(magicHeader)+nonceSize:]

	// 6. Create AES-256-GCM cipher
	block, err := aes.NewCipher(decryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher for decryption: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher for decryption: %w", err)
	}

	// 7. Decrypt the message
	decryptedMessage, err := gcm.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt message: %w", err)
	}

	return string(decryptedMessage), nil
}