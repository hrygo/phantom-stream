package injector

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
