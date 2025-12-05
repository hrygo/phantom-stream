//go:build integration
// +build integration

package injector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Test constants for Phase 7
// testPDFPath is defined in integration_test.go

// TestWatermarkDualAnchorSign tests the dual-anchor signature mechanism
func TestWatermarkDualAnchorSign(t *testing.T) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping Phase 7 integration test")
	}

	tests := []struct {
		name        string
		message     string
		key         string
		expectError bool
	}{
		{
			name:        "Valid dual-anchor signature",
			message:     "WatermarkDualAnchor:Test-DualAnchor",
			key:         testKey32,
			expectError: false,
		},
		{
			name:        "Long message",
			message:     "WatermarkDualAnchor:VeryLongMessage-" + string(make([]byte, 100)),
			key:         testKey32,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test signing
			err := Sign(testPDFPath, tt.message, tt.key, nil)

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
			expectedPath := testPDFPath[:len(testPDFPath)-4] + "_signed.pdf"
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Errorf("Signed file not created: %s", expectedPath)
				return
			}

			// Clean up
			defer os.Remove(expectedPath)
		})
	}
}

// TestWatermarkDualAnchorVerify tests verification with both anchors
func TestWatermarkDualAnchorVerify(t *testing.T) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping Phase 7 integration test")
	}

	testMessage := "WatermarkDualAnchor:Verify-Test"

	// Create signed PDF
	err := Sign(testPDFPath, testMessage, testKey32, nil)
	if err != nil {
		t.Fatalf("Failed to create test signed PDF: %v", err)
	}

	signedPath := testPDFPath[:len(testPDFPath)-4] + "_signed.pdf"
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
			message, _, err := Verify(signedPath, tt.key, nil)

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

// TestWatermarkDualAnchorSMaskAnchorFallback tests SMask anchor as fallback
func TestWatermarkDualAnchorSMaskAnchorFallback(t *testing.T) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping Phase 7 integration test")
	}

	testMessage := "WatermarkDualAnchor:SMask-Fallback-Test"

	// Create signed PDF
	err := Sign(testPDFPath, testMessage, testKey32, nil)
	if err != nil {
		t.Fatalf("Failed to create test signed PDF: %v", err)
	}

	signedPath := testPDFPath[:len(testPDFPath)-4] + "_signed.pdf"
	defer os.Remove(signedPath)

	// Remove attachment to test SMask fallback
	noAttachPath := filepath.Join(t.TempDir(), "test_noattach.pdf")

	conf := model.NewDefaultConfiguration()
	err = api.RemoveAttachmentsFile(signedPath, noAttachPath, nil, conf)
	if err != nil {
		t.Fatalf("Failed to remove attachments: %v", err)
	}

	// Verify using SMask anchor only
	message, _, err := Verify(noAttachPath, testKey32, nil)
	if err != nil {
		t.Fatalf("SMask fallback verification failed: %v", err)
	}

	if message != testMessage {
		t.Errorf("SMask fallback message mismatch: got '%s', want '%s'", message, testMessage)
	}
}

// TestWatermarkDualAnchorAttachmentAnchorOnly tests attachment anchor independently
func TestWatermarkDualAnchorAttachmentAnchorOnly(t *testing.T) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping Phase 7 integration test")
	}

	testMessage := "WatermarkDualAnchor:Attachment-Only-Test"

	// Create signed PDF
	err := Sign(testPDFPath, testMessage, testKey32, nil)
	if err != nil {
		t.Fatalf("Failed to create test signed PDF: %v", err)
	}

	signedPath := testPDFPath[:len(testPDFPath)-4] + "_signed.pdf"
	defer os.Remove(signedPath)

	// Verify using attachment anchor (should succeed first)
	message, _, err := Verify(signedPath, testKey32, nil)
	if err != nil {
		t.Fatalf("Attachment anchor verification failed: %v", err)
	}

	if message != testMessage {
		t.Errorf("Attachment anchor message mismatch: got '%s', want '%s'", message, testMessage)
	}
}

// TestWatermarkDualAnchorSMaskInjection tests SMask injection mechanism
func TestWatermarkDualAnchorSMaskInjection(t *testing.T) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping Phase 7 SMask test")
	}

	outputPath := filepath.Join(t.TempDir(), "test_smask.pdf")

	// Create test payload
	testPayload, err := createEncryptedPayload("WatermarkDualAnchor:SMask-Test", []byte(testKey32))
	if err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
	}

	// Test SMask injection
	err = InjectSMaskAnchor(testPDFPath, outputPath, testPayload)
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
	payload, err := ExtractSMaskPayloadFromPDF(outputPath)
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

// TestWatermarkDualAnchorFileSizeImpact tests file size changes after signing
func TestWatermarkDualAnchorFileSizeImpact(t *testing.T) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping Phase 7 file size test")
	}

	// Get original file size
	origInfo, err := os.Stat(testPDFPath)
	if err != nil {
		t.Fatalf("Failed to stat original file: %v", err)
	}
	origSize := origInfo.Size()

	// Create signed PDF with lightweight anchors (Attachment + Content)
	// to test reasonable file size impact
	err = Sign(testPDFPath, "WatermarkDualAnchor:Size-Test", testKey32, []string{"Attachment", "Content"})
	if err != nil {
		t.Fatalf("Failed to sign PDF: %v", err)
	}

	signedPath := testPDFPath[:len(testPDFPath)-4] + "_signed.pdf"
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

	// File size should not increase dramatically with lightweight anchors
	// Attachment + Content should be < 5% increase
	if pct > 5.0 {
		t.Errorf("File size increased too much: %.2f%% (expected < 5%% for lightweight anchors)", pct)
	}
}

// BenchmarkWatermarkDualAnchorSign benchmarks the dual-anchor signing process
func BenchmarkWatermarkDualAnchorSign(b *testing.B) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		b.Skip("Test PDF not found, skipping Phase 7 benchmark")
	}

	testMessage := "WatermarkDualAnchor:Benchmark-Test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Sign(testPDFPath, testMessage, testKey32, nil)
		if err != nil {
			b.Fatalf("Sign failed: %v", err)
		}

		// Clean up
		signedPath := testPDFPath[:len(testPDFPath)-4] + "_signed.pdf"
		os.Remove(signedPath)
	}
}

// BenchmarkWatermarkDualAnchorVerify benchmarks the dual-anchor verification process
func BenchmarkWatermarkDualAnchorVerify(b *testing.B) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		b.Skip("Test PDF not found, skipping Phase 7 benchmark")
	}

	testMessage := "WatermarkDualAnchor:Benchmark-Verify-Test"

	// Create signed PDF once
	err := Sign(testPDFPath, testMessage, testKey32, nil)
	if err != nil {
		b.Fatalf("Failed to create test signed PDF: %v", err)
	}

	signedPath := testPDFPath[:len(testPDFPath)-4] + "_signed.pdf"
	defer os.Remove(signedPath)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := Verify(signedPath, testKey32, nil)
		if err != nil {
			b.Fatalf("Verify failed: %v", err)
		}
	}
}
