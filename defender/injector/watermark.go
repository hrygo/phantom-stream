package injector

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	// "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model" // Unused
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

var (
	magicHeader = []byte{0xCA, 0xFE, 0xBA, 0xBE} // Magic Header for identifying hidden data
	keySize     = 32                               // AES-256 key size in bytes
	nonceSize   = 12                               // GCM standard nonce size
	metaKey     = "PStream"                        // Key for custom metadata
)

// Sign embeds an encrypted message into PDF metadata by directly modifying the Info dictionary.
func Sign(filePath, message, key string) error {
	if len(key) != keySize {
		return fmt.Errorf("encryption key must be %d bytes long", keySize)
	}
	encryptionKey := []byte(key)

	// 1. Create AES-256-GCM cipher
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// 2. Generate random Nonce
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// 3. Encrypt the message
	encryptedMessage := gcm.Seal(nil, nonce, []byte(message), nil)

	// 4. Build Raw Payload: Magic Header + Nonce + Encrypted Message
	rawPayload := make([]byte, 0, len(magicHeader)+len(nonce)+len(encryptedMessage))
	rawPayload = append(rawPayload, magicHeader...)
	rawPayload = append(rawPayload, nonce...)
	rawPayload = append(rawPayload, encryptedMessage...)

	// 5. Encode Payload to Hex String
	payloadHex := hex.EncodeToString(rawPayload)

	// 6. Load PDF Context
	ctx, err := api.ReadContextFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read PDF context: %w", err)
	}

	// 7. Inject Metadata into Info Dictionary
	var infoDict types.Dict

	if ctx.Info == nil {
		// Create new Info dictionary
		infoDict = types.NewDict()
		indRef, err := ctx.XRefTable.IndRefForNewObject(infoDict)
		if err != nil {
			return fmt.Errorf("failed to create new object for Info dict: %w", err)
		}
		ctx.Info = indRef
	} else {
		// Dereference existing Info dictionary
		obj, err := ctx.XRefTable.Dereference(ctx.Info)
		if err != nil {
			return fmt.Errorf("failed to dereference Info dict: %w", err)
		}
		// Resolve any further indirect references
		for {
			if indRef, ok := obj.(*types.IndirectRef); ok {
				obj, err = ctx.XRefTable.Dereference(*indRef)
				if err != nil {
					return fmt.Errorf("failed to dereference Info dict chain: %w", err)
				}
			} else {
				break
			}
		}
		var ok bool
		infoDict, ok = obj.(types.Dict)
		if !ok {
			return fmt.Errorf("Info object is not a dictionary, it is %T", obj)
		}
	}

	// Set custom property
	infoDict[metaKey] = types.StringLiteral(payloadHex)

	fmt.Printf("DEBUG: Injected metadata key '%s' with payload length %d\n", metaKey, len(payloadHex))

	// 8. Write modified PDF to a new file (WriteContextFile creates a fresh PDF, effectively Optimizing)
	outputFileName := fmt.Sprintf("%s_signed.pdf", filepath.Base(filePath)[:len(filepath.Base(filePath))-len(filepath.Ext(filePath))])
	outputFilePath := filepath.Join(filepath.Dir(filePath), outputFileName)

	if err := api.WriteContextFile(ctx, outputFilePath); err != nil {
		return fmt.Errorf("failed to write modified PDF: %w", err)
	}

	return nil
}

// Verify extracts and verifies a hidden message from PDF metadata using manual context inspection.
func Verify(filePath, key string) (string, error) {
	if len(key) != keySize {
		return "", fmt.Errorf("decryption key must be %d bytes long", keySize)
	}
	decryptionKey := []byte(key)

	// 1. Read Context
	ctx, err := api.ReadContextFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read PDF context: %w", err)
	}

	// 2. Extract Metadata from Info Dictionary
	var payloadHex string
	
	if ctx.Info != nil {
		// Dereference the Info object
		obj, err := ctx.XRefTable.Dereference(ctx.Info)
		if err != nil {
			return "", fmt.Errorf("failed to dereference Info dict: %w", err)
		}
		
		// Resolve any further indirect references
		for {
			if indRef, ok := obj.(*types.IndirectRef); ok {
				obj, err = ctx.XRefTable.Dereference(*indRef)
				if err != nil {
					return "", fmt.Errorf("failed to dereference Info dict chain: %w", err)
				}
			} else {
				break
			}
		}

		if infoDict, ok := obj.(types.Dict); ok {
			// Look for our key
			if val, ok := infoDict[metaKey]; ok {
				// Let's try casting to specific types for cleaner extraction
				switch v := val.(type) {
				case types.StringLiteral:
					payloadHex = v.Value()
				case types.HexLiteral:
					payloadHex = v.Value()
				default:
					// Fallback to String() and strip delimiters
					s := val.String()
					if len(s) > 2 && s[0] == '(' && s[len(s)-1] == ')' {
						payloadHex = s[1 : len(s)-1]
					} else if len(s) > 2 && s[0] == '<' && s[len(s)-1] == '>' {
						payloadHex = s[1 : len(s)-1]
					} else {
						payloadHex = s
					}
				}
			}
		}
	}

	if payloadHex == "" {
		return "", errors.New("metadata key not found")
	}

	// fmt.Printf("DEBUG: Found metadata raw: %s\n", payloadHex)

	// 3. Decode Hex
	rawPayload, err := hex.DecodeString(payloadHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex payload: %w", err)
	}

	// 4. Verify Magic Header
	if len(rawPayload) < len(magicHeader) {
		return "", errors.New("payload too short for magic header")
	}
	
	for i := range magicHeader {
		if rawPayload[i] != magicHeader[i] {
			return "", errors.New("magic header mismatch")
		}
	}

	// 5. Extract Nonce and Ciphertext
	if len(rawPayload) < len(magicHeader)+nonceSize {
		return "", errors.New("payload too short")
	}

	nonce := rawPayload[len(magicHeader) : len(magicHeader)+nonceSize]
	encryptedMessage := rawPayload[len(magicHeader)+nonceSize:]

	// 6. Decrypt
	block, err := aes.NewCipher(decryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	decryptedMessage, err := gcm.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt message: %w", err)
	}

	return string(decryptedMessage), nil
}
