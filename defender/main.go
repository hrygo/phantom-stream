package main

import (
	"fmt"
	"os"
	"strings"

	"defender/injector"

	"github.com/spf13/cobra"
)

var (
	filePath   string
	message    string
	key        string
	round      string
	verifyMode string
	version    = "1.0.0"
)

var rootCmd = &cobra.Command{
	Use:   "phantom-guard",
	Short: "PhantomGuard - PDF watermark embedding and verification tool",
	Long: `PhantomGuard is a CLI tool for embedding encrypted tracking information 
into PDF files without disrupting the reading experience.

This tool is part of the PhantomStream defense system and uses 
PDF embedded attachments to ensure tracking information survives 
advanced cleaning attacks.

Version: ` + version,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		runInteractive()
	},
}

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Embed encrypted tracking message into a PDF file",
	Long: `The sign command embeds an encrypted message (e.g., employee ID, 
tracking code) into a PDF file as a hidden attachment.

The original PDF remains fully readable, and the tracking information 
can only be extracted with the correct decryption key.

Example:
  defender sign -f report.pdf -m "UserID:12345" -k "MySecretKey32BytesLongString!!"

Note: The encryption key must be exactly 32 bytes long.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		if filePath == "" {
			return fmt.Errorf("required flag --file is missing")
		}
		if message == "" {
			return fmt.Errorf("required flag --msg is missing")
		}
		if key == "" {
			return fmt.Errorf("required flag --key is missing")
		}

		fmt.Printf("🛡️  Defender Sign Operation\n")
		fmt.Printf("   File: %s\n", filePath)
		fmt.Printf("   Message: %s\n", message)
		fmt.Println()

		err := injector.Sign(filePath, message, key, round, nil)
		if err != nil {
			return fmt.Errorf("sign operation failed: %w", err)
		}

		fmt.Println("\n✅ Sign operation completed successfully!")
		return nil
	},
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Extract and verify hidden tracking message from a PDF file",
	Long: `The verify command extracts the embedded tracking message from a 
signed PDF file and decrypts it using the provided key.

This operation will fail if:
  - The file does not contain an embedded tracking message
  - The decryption key is incorrect
  - The file has been cleaned or modified

Example:
  defender verify -f report_signed.pdf -k "MySecretKey32BytesLongString!!"

Note: The decryption key must match the one used during signing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		if filePath == "" {
			return fmt.Errorf("required flag --file is missing")
		}
		if key == "" {
			return fmt.Errorf("required flag --key is missing")
		}

		fmt.Printf("🔍 Defender Verify Operation\n")
		fmt.Printf("   File: %s\n", filePath)
		fmt.Println()

		if strings.ToLower(verifyMode) == "all" {
			crypto, err := injector.NewCryptoManager([]byte(key))
			if err != nil {
				return fmt.Errorf("failed to create crypto manager: %w", err)
			}
			registry := injector.NewAnchorRegistry()
			anchors := registry.GetAvailableAnchors()
			anySuccess := false
			for _, a := range anchors {
				if a.Name() == "Visual" { // Visual 不支持提取
					continue
				}
				fmt.Printf(" - Trying %s... ", a.Name())
				payload, extErr := a.Extract(filePath)
				if extErr != nil {
					fmt.Println("extract failed")
					continue
				}
				msg, decErr := crypto.Decrypt(payload)
				if decErr != nil {
					fmt.Println("decrypt failed")
					continue
				}
				fmt.Println("OK")
				fmt.Printf("   Message(%s): %s\n", a.Name(), msg)
				anySuccess = true
			}
			if !anySuccess {
				return fmt.Errorf("verify operation failed: all anchors invalid or missing")
			}
			fmt.Println("✅ Verification finished (mode=all).")
			return nil
		}

		extractedMsg, _, err := injector.Verify(filePath, key, nil)
		if err != nil {
			return fmt.Errorf("verify operation failed: %w", err)
		}

		fmt.Println("✅ Verification successful!")
		fmt.Printf("📋 Extracted message: \"%s\"\n", extractedMsg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(verifyCmd)

	// Sign command flags
	signCmd.Flags().StringVarP(&filePath, "file", "f", "", "Source PDF file path (required)")
	signCmd.Flags().StringVarP(&message, "msg", "m", "", "Message to embed, e.g., 'UserID:123' (required)")
	signCmd.Flags().StringVarP(&key, "key", "k", "", "32-byte encryption key (required)")
	signCmd.Flags().StringVarP(&round, "round", "r", "", "Round tag appended to output filename (optional), e.g., '11' -> *_11_signed.pdf")
	_ = signCmd.MarkFlagRequired("file")
	_ = signCmd.MarkFlagRequired("msg")
	_ = signCmd.MarkFlagRequired("key")

	// Verify command flags
	verifyCmd.Flags().StringVarP(&filePath, "file", "f", "", "Target PDF file path (required)")
	verifyCmd.Flags().StringVarP(&key, "key", "k", "", "32-byte decryption key (required)")
	verifyCmd.Flags().StringVar(&verifyMode, "mode", "auto", "Verification mode: auto|all")
	_ = verifyCmd.MarkFlagRequired("file")
	_ = verifyCmd.MarkFlagRequired("key")
}

func Execute() error {
	return rootCmd.Execute()
}

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		os.Exit(1)
	}
}
