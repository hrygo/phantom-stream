package injector

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestAnchorOverhead(t *testing.T) {
	// Setup paths
	inputPath := "../testdata/2511.17467v2.pdf"

	// Ensure input exists
	info, err := os.Stat(inputPath)
	if err != nil {
		t.Fatalf("Input file not found: %v", err)
	}
	originalSize := info.Size()

	// Define payload
	message := "BenchmarkPayload12345"
	key := testKey32

	// Test cases
	tests := []struct {
		name    string
		anchors []string
	}{
		{"Attachment", []string{"Attachment"}},
		{"SMask", []string{"SMask"}},
		{"Content", []string{"Content"}},
		{"Visual", []string{"Visual"}},
		{"All Combined", []string{"Attachment", "SMask", "Content", "Visual"}},
	}

	fmt.Printf("\n=== Anchor Overhead Benchmark ===\n")
	fmt.Printf("Original File Size: %.2f KB (%d bytes)\n", float64(originalSize)/1024, originalSize)
	fmt.Printf("%-20s | %-15s | %-15s\n", "Strategy", "Result Size", "Overhead")
	fmt.Printf("---------------------|-----------------|----------------\n")

	for _, tt := range tests {
		safeName := tt.name
		if safeName == "All Combined" {
			safeName = "All_Combined"
		}
		outputPath := filepath.Join("..", "testdata", fmt.Sprintf("bench_%s.pdf", safeName))

		// Sign creates a file with _signed suffix, we need to move it to our benchmark path
		// Reconstruct default signed path
		baseName := filepath.Base(inputPath)
		baseNameNoExt := baseName[:len(baseName)-len(filepath.Ext(baseName))]
		defaultSignedPath := filepath.Join(filepath.Dir(inputPath), baseNameNoExt+"_signed"+filepath.Ext(inputPath))

		err := Sign(inputPath, message, key, tt.anchors)
		if err != nil {
			t.Errorf("Failed to sign %s: %v", tt.name, err)
			continue
		}

		// Move to expected benchmark path
		err = os.Rename(defaultSignedPath, outputPath)
		if err != nil {
			t.Errorf("Failed to move signed file for %s: %v", tt.name, err)
			continue
		}

		// Measure
		outInfo, err := os.Stat(outputPath)
		if err != nil {
			t.Errorf("Failed to stat output %s: %v", tt.name, err)
			continue
		}

		newSize := outInfo.Size()
		overhead := newSize - originalSize

		fmt.Printf("%-20s | %.2f KB        | %+d bytes\n",
			tt.name,
			float64(newSize)/1024,
			overhead,
		)

		// Verify (sanity check)
		if tt.name != "Visual" && tt.name != "All Combined" {
			_, _, err := Verify(outputPath, key, nil)
			if err != nil {
				t.Errorf("Verification failed for %s: %v", tt.name, err)
			}
		}

		// Cleanup
		os.Remove(outputPath)
	}
	fmt.Println()
}
