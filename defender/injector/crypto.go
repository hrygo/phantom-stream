package injector

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// CryptoManager handles encryption and decryption operations
type CryptoManager struct {
	key []byte
}

// NewCryptoManager creates a new crypto manager with the given key
func NewCryptoManager(key []byte) (*CryptoManager, error) {
	if len(key) != keySize {
		return nil, ErrInvalidKeySize
	}
	return &CryptoManager{key: key}, nil
}

// Encrypt encrypts a message and returns the encrypted payload
// Payload format: magic header + nonce + encrypted message
func (c *CryptoManager) Encrypt(message string) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
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

// Decrypt decrypts a payload and returns the original message
func (c *CryptoManager) Decrypt(payload []byte) (string, error) {
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
	block, err := aes.NewCipher(c.key)
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
