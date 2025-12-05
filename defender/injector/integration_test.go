//go:build integration
// +build integration

package injector

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	testPDFPath = "../testdata/2511.17467v2.pdf"
)

// TestAnchorCombinations tests various combinations of anchors
func TestAnchorCombinations(t *testing.T) {
	// Skip if test PDF doesn't exist
	if _, err := os.Stat(testPDFPath); os.IsNotExist(err) {
		t.Skip("Test PDF not found, skipping integration test")
	}

	testMessage := "IntegrationTest:Combinations"
	testKey := testKey32

	tests := []struct {
		name            string
		selectedAnchors []string
		expectVerify    bool     // Whether verification should succeed
		verifyAnchors   []string // Anchors to verify with (nil for auto)
		expectedAnchor  string   // Expected anchor name to be found (if any specific one is targeted)
	}{
		{
			name:            "Standard (Attachment Only)",
			selectedAnchors: []string{"Attachment"},
			expectVerify:    true,
			verifyAnchors:   nil,
			expectedAnchor:  "Attachment",
		},
		{
			name:            "Stealth (Attachment + SMask)",
			selectedAnchors: []string{"Attachment", "SMask"},
			expectVerify:    true,
			verifyAnchors:   nil,
			expectedAnchor:  "Attachment", // First one found
		},
		{
			name:            "Content Only",
			selectedAnchors: []string{"Content"},
			expectVerify:    true,
			verifyAnchors:   nil,
			expectedAnchor:  "Content",
		},
		{
			name:            "Visual Only (Should fail verify)",
			selectedAnchors: []string{"Visual"},
			expectVerify:    false, // Visual is not extractable
			verifyAnchors:   nil,
		},
		{
			name:            "Custom: SMask + Content",
			selectedAnchors: []string{"SMask", "Content"},
			expectVerify:    true,
			verifyAnchors:   nil,
			expectedAnchor:  "SMask", // First one found
		},
		{
			name:            "Verify Specific: Content",
			selectedAnchors: []string{"Attachment", "Content"},
			expectVerify:    true,
			verifyAnchors:   []string{"Content"},
			expectedAnchor:  "Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			round := "ComboTest"

			// 1. Sign
			err := Sign(testPDFPath, testMessage, testKey, round, tt.selectedAnchors)
			if err != nil {
				t.Fatalf("Sign failed: %v", err)
			}

			// Determine signed file path
			dir := filepath.Dir(testPDFPath)
			base := filepath.Base(testPDFPath)
			ext := filepath.Ext(base)
			name := base[:len(base)-len(ext)]
			signedPath := filepath.Join(dir, name+"_"+round+"_signed"+ext)

			defer os.Remove(signedPath)

			// 2. Verify
			msg, anchor, err := Verify(signedPath, testKey, tt.verifyAnchors)

			if tt.expectVerify {
				if err != nil {
					t.Fatalf("Verify failed unexpectedly: %v", err)
				}
				if msg != testMessage {
					t.Errorf("Message mismatch: got '%s', want '%s'", msg, testMessage)
				}
				if tt.expectedAnchor != "" && anchor != tt.expectedAnchor {
					t.Errorf("Anchor mismatch: got '%s', want '%s'", anchor, tt.expectedAnchor)
				}
			} else {
				if err == nil {
					t.Error("Expected verify error, got nil")
				}
			}
		})
	}
}
