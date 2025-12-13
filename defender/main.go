package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"defender/injector"

	"github.com/spf13/cobra"
)

var (
	filePath   string
	message    string
	key        string
	verifyMode string
	version    = "1.2.0"
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

		// Handle Key (Flag -> Env -> Error)
		if key == "" {
			key = os.Getenv("DEFAULT_KEY")
			if key == "" {
				return fmt.Errorf("required flag --key is missing and DEFAULT_KEY env not set")
			}
			fmt.Println("ℹ️  Using key from environment variable DEFAULT_KEY")
		}

		fmt.Printf("🛡️  Defender Sign Operation\n")
		fmt.Printf("   File: %s\n", filePath)
		fmt.Printf("   Message: %s\n", message)
		fmt.Println()

		err := injector.Sign(filePath, message, key, nil)
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

		// Handle Key (Flag -> Env -> Error)
		if key == "" {
			key = os.Getenv("DEFAULT_KEY")
			if key == "" {
				return fmt.Errorf("required flag --key is missing and DEFAULT_KEY env not set")
			}
			fmt.Println("ℹ️  Using key from environment variable DEFAULT_KEY")
		}

		fmt.Printf("🔍 Defender Verify Operation\n")
		fmt.Printf("   File: %s\n", filePath)
		fmt.Println()

		if strings.EqualFold(verifyMode, "all") {
			crypto, err := injector.NewCryptoManager([]byte(key))
			if err != nil {
				return fmt.Errorf("failed to create crypto manager: %w", err)
			}
			registry := injector.NewAnchorRegistry()
			anchors := registry.GetAvailableAnchors()
			anySuccess := false
			for _, a := range anchors {
				if a.Name() == injector.AnchorNameVisual { // Visual 不支持提取
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

var initKeyCmd = &cobra.Command{
	Use:   "init-key",
	Short: "Generate initialization key to .env file (silent)",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Generate Key (32 chars)
		k := make([]byte, 16)
		if _, err := rand.Read(k); err != nil {
			return err
		}
		keyVal := hex.EncodeToString(k)

		// 2. Determine path (Binary directory)
		exePath, err := os.Executable()
		if err != nil {
			return err
		}
		envPath := filepath.Join(filepath.Dir(exePath), ".env")

		// 3. Read/Update .env
		contentByte, _ := os.ReadFile(envPath)
		content := string(contentByte)
		newLine := fmt.Sprintf("DEFAULT_KEY=%s", keyVal)

		if strings.Contains(content, "DEFAULT_KEY=") {
			// Replace existing key
			lines := strings.Split(content, "\n")
			for i, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "DEFAULT_KEY=") {
					oldVal := strings.TrimPrefix(strings.TrimSpace(line), "DEFAULT_KEY=")
					fmt.Printf("ℹ️  Overwriting old key: %s\n", oldVal)
					lines[i] = newLine
				}
			}
			content = strings.Join(lines, "\n")
		} else {
			// Append new key
			if len(content) > 0 && !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += newLine + "\n"
		}

		return os.WriteFile(envPath, []byte(content), 0644)
	},
}

// setupCommands initializes command line flags
func setupCommands() {
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(initKeyCmd)

	// Sign command flags
	signCmd.Flags().StringVarP(&filePath, "file", "f", "", "Source PDF file path (required)")
	signCmd.Flags().StringVarP(&message, "msg", "m", "", "Message to embed, e.g., 'UserID:123' (required)")
	signCmd.Flags().StringVarP(&key, "key", "k", "", "32-byte encryption key (optional if DEFAULT_KEY env is set)")
	_ = signCmd.MarkFlagRequired("file")
	_ = signCmd.MarkFlagRequired("msg")

	// Verify command flags
	verifyCmd.Flags().StringVarP(&filePath, "file", "f", "", "Target PDF file path (required)")
	verifyCmd.Flags().StringVarP(&key, "key", "k", "", "32-byte decryption key (optional if DEFAULT_KEY env is set)")
	verifyCmd.Flags().StringVar(&verifyMode, "mode", "auto", "Verification mode: auto|all")
	_ = verifyCmd.MarkFlagRequired("file")
}

func Execute() error {
	return rootCmd.Execute()
}

// loadEnv loads environment variables from .env file in the binary's directory
func loadEnv() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	envPath := filepath.Join(filepath.Dir(exePath), ".env")

	data, err := os.ReadFile(envPath)
	if err != nil {
		return // No .env found, ignore
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Only set if not already present in environment
			// This allows manual override via export VAR=...
			if _, exists := os.LookupEnv(key); !exists {
				os.Setenv(key, val)
			}
		}
	}
}

func main() {
	loadEnv()
	setupCommands()
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ Error: %v\n", err)
		os.Exit(1)
	}
}
