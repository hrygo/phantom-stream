package injector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Test constants for Phase 7
const (
	testPDFPath = "../../docs/2511.17467v2.pdf"
)

// TestPhase7DualAnchorSign tests the dual-anchor signature mechanism
func TestPhase7DualAnchorSign(t *testing.T) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping Phase 7 integration test")
	}

	tests := []struct {
		name        string
		message     string
		key         string
		round       string
		expectError bool
	}{
		{
			name:        "Valid dual-anchor signature",
			message:     "Phase7:Test-DualAnchor",
			key:         testKey32,
			round:       "Test",
			expectError: false,
		},
		{
			name:        "Long message",
			message:     "Phase7:VeryLongMessage-" + string(make([]byte, 100)),
			key:         testKey32,
			round:       "Test",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test signing
			err := Sign(testPDFPath, tt.message, tt.key, tt.round)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Sign failed: %v", err)
			}

			// Verify signed file exists
			expectedPath := testPDFPath[:len(testPDFPath)-4] + "_" + tt.round + "_signed.pdf"
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Errorf("Signed file not created: %s", expectedPath)
				return
			}

			// Clean up
			defer os.Remove(expectedPath)
		})
	}
}

// TestPhase7DualAnchorVerify tests verification with both anchors
func TestPhase7DualAnchorVerify(t *testing.T) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping Phase 7 integration test")
	}

	testMessage := "Phase7:Verify-Test"
	testRound := "VerifyTest"

	// Create signed PDF
	err := Sign(testPDFPath, testMessage, testKey32, testRound)
	if err != nil {
		t.Fatalf("Failed to create test signed PDF: %v", err)
	}

	signedPath := testPDFPath[:len(testPDFPath)-4] + "_" + testRound + "_signed.pdf"
	defer os.Remove(signedPath)

	tests := []struct {
		name        string
		key         string
		expectError bool
		expectMsg   string
	}{
		{
			name:        "Valid verification with correct key",
			key:         testKey32,
			expectError: false,
			expectMsg:   testMessage,
		},
		{
			name:        "Invalid verification with wrong key",
			key:         "wrongkey1234567890123456789012",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, err := Verify(signedPath, tt.key)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Verify failed: %v", err)
			}

			if message != tt.expectMsg {
				t.Errorf("Message mismatch: got '%s', want '%s'", message, tt.expectMsg)
			}
		})
	}
}

// TestPhase7SMaskAnchorFallback tests SMask anchor as fallback
func TestPhase7SMaskAnchorFallback(t *testing.T) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping Phase 7 integration test")
	}

	testMessage := "Phase7:SMask-Fallback-Test"
	testRound := "SMaskTest"

	// Create signed PDF
	err := Sign(testPDFPath, testMessage, testKey32, testRound)
	if err != nil {
		t.Fatalf("Failed to create test signed PDF: %v", err)
	}

	signedPath := testPDFPath[:len(testPDFPath)-4] + "_" + testRound + "_signed.pdf"
	defer os.Remove(signedPath)

	// Remove attachment to test SMask fallback
	noAttachPath := filepath.Join(t.TempDir(), "test_noattach.pdf")

	conf := model.NewDefaultConfiguration()
	err = api.RemoveAttachmentsFile(signedPath, noAttachPath, nil, conf)
	if err != nil {
		t.Fatalf("Failed to remove attachments: %v", err)
	}

	// Verify using SMask anchor only
	message, err := Verify(noAttachPath, testKey32)
	if err != nil {
		t.Fatalf("SMask fallback verification failed: %v", err)
	}

	if message != testMessage {
		t.Errorf("SMask fallback message mismatch: got '%s', want '%s'", message, testMessage)
	}
}

// TestPhase7AttachmentAnchorOnly tests attachment anchor independently
func TestPhase7AttachmentAnchorOnly(t *testing.T) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping Phase 7 integration test")
	}

	testMessage := "Phase7:Attachment-Only-Test"
	testRound := "AttachTest"

	// Create signed PDF
	err := Sign(testPDFPath, testMessage, testKey32, testRound)
	if err != nil {
		t.Fatalf("Failed to create test signed PDF: %v", err)
	}

	signedPath := testPDFPath[:len(testPDFPath)-4] + "_" + testRound + "_signed.pdf"
	defer os.Remove(signedPath)

	// Verify using attachment anchor (should succeed first)
	message, err := Verify(signedPath, testKey32)
	if err != nil {
		t.Fatalf("Attachment anchor verification failed: %v", err)
	}

	if message != testMessage {
		t.Errorf("Attachment anchor message mismatch: got '%s', want '%s'", message, testMessage)
	}
}

// TestPhase7SMaskInjection tests SMask injection mechanism
func TestPhase7SMaskInjection(t *testing.T) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping Phase 7 SMask test")
	}

	outputPath := filepath.Join(t.TempDir(), "test_smask.pdf")

	// Create test payload
	testPayload, err := createEncryptedPayload("Phase7:SMask-Test", []byte(testKey32))
	if err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
	}

	// Test SMask injection
	err = injectSMaskAnchor(testPDFPath, outputPath, testPayload)
	if err != nil {
		// If PDF has no images, expect error
		if err.Error() != "no images found in PDF (SMask anchor requires at least one image)" {
			t.Fatalf("SMask injection failed unexpectedly: %v", err)
		}
		t.Skip("PDF has no images, skipping SMask injection test")
		return
	}

	// Verify output file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("SMask-injected PDF not created")
	}

	// Try to extract payload
	payload, err := extractSMaskPayloadFromPDF(outputPath)
	if err != nil {
		t.Fatalf("Failed to extract SMask payload: %v", err)
	}

	if len(payload) == 0 {
		t.Error("Extracted payload is empty")
	}

	// Verify payload structure
	if len(payload) < len(magicHeader) {
		t.Error("Extracted payload too short to contain magic header")
	}
}

// TestPhase7FileSizeImpact tests file size changes after signing
func TestPhase7FileSizeImpact(t *testing.T) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping Phase 7 file size test")
	}

	testRound := "SizeTest"

	// Get original file size
	origInfo, err := os.Stat(testPDFPath)
	if err != nil {
		t.Fatalf("Failed to stat original file: %v", err)
	}
	origSize := origInfo.Size()

	// Create signed PDF
	err = Sign(testPDFPath, "Phase7:Size-Test", testKey32, testRound)
	if err != nil {
		t.Fatalf("Failed to sign PDF: %v", err)
	}

	signedPath := testPDFPath[:len(testPDFPath)-4] + "_" + testRound + "_signed.pdf"
	defer os.Remove(signedPath)

	// Get signed file size
	signedInfo, err := os.Stat(signedPath)
	if err != nil {
		t.Fatalf("Failed to stat signed file: %v", err)
	}
	signedSize := signedInfo.Size()

	// Calculate size difference
	diff := signedSize - origSize
	pct := float64(diff) / float64(origSize) * 100

	t.Logf("Original size: %d bytes", origSize)
	t.Logf("Signed size: %d bytes", signedSize)
	t.Logf("Size difference: %d bytes (%.2f%%)", diff, pct)

	// File size should not increase dramatically (< 1%)
	if pct > 1.0 {
		t.Errorf("File size increased too much: %.2f%% (expected < 1%%)", pct)
	}
}

// BenchmarkPhase7DualAnchorSign benchmarks the dual-anchor signing process
func BenchmarkPhase7DualAnchorSign(b *testing.B) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		b.Skip("Test PDF not found, skipping Phase 7 benchmark")
	}

	testMessage := "Phase7:Benchmark-Test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testRound := "Bench" + string(rune(i))
		err := Sign(testPDFPath, testMessage, testKey32, testRound)
		if err != nil {
			b.Fatalf("Sign failed: %v", err)
		}

		// Clean up
		signedPath := testPDFPath[:len(testPDFPath)-4] + "_" + testRound + "_signed.pdf"
		os.Remove(signedPath)
	}
}

// BenchmarkPhase7DualAnchorVerify benchmarks the dual-anchor verification process
func BenchmarkPhase7DualAnchorVerify(b *testing.B) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		b.Skip("Test PDF not found, skipping Phase 7 benchmark")
	}

	testMessage := "Phase7:Benchmark-Verify-Test"
	testRound := "BenchVerify"

	// Create signed PDF once
	err := Sign(testPDFPath, testMessage, testKey32, testRound)
	if err != nil {
		b.Fatalf("Failed to create test signed PDF: %v", err)
	}

	signedPath := testPDFPath[:len(testPDFPath)-4] + "_" + testRound + "_signed.pdf"
	defer os.Remove(signedPath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Verify(signedPath, testKey32)
		if err != nil {
			b.Fatalf("Verify failed: %v", err)
		}
	}
}
