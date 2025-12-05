package injector

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

//go:embed assets/GoNotoCurrent-Regular.ttf
var goNotoCurrentTTF []byte

// InstallEmbeddedUnicodeFont installs the embedded pan-Unicode font into pdfcpu's user font registry.
// It writes the embedded TTF to a temporary location and calls api.InstallFonts.
//
// IMPORTANT NOTES:
//  1. The font file "GoNotoCurrent-Regular.ttf" is from Go Noto Universal project (v7.0)
//     Source: https://github.com/satbyy/go-noto-universal
//  2. After installation, pdfcpu registers the font with name: "GoNotoCurrent-Regular-Regular"
//     (The double "Regular" comes from TTF internal metadata, see anchor_visual.go for details)
//  3. This font supports the entire Unicode BMP + supplementary planes, covering:
//     - CJK (Chinese, Japanese, Korean)
//     - Arabic, Cyrillic, Hebrew, Thai, etc.
//     - Emoji and symbols
//  4. Font size: ~14MB embedded in binary
//  5. Installation happens once per program run (protected by fontMutex in anchor_visual.go)
func InstallEmbeddedUnicodeFont() error {
	tmpDir, err := os.MkdirTemp("", "phantom-unicode-font")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	ttfPath := filepath.Join(tmpDir, "GoNotoCurrent-Regular.ttf")
	if err := os.WriteFile(ttfPath, goNotoCurrentTTF, 0644); err != nil {
		return err
	}
	return api.InstallFonts([]string{ttfPath})
}
