package main

import (
	"fmt"
	"log"
	"os"

	"defender/injector"

	"github.com/spf13/cobra"
)

var (
	filePath string
	message  string
	key      string
)

var rootCmd = &cobra.Command{
	Use:   "defender",
	Short: "Defender is a tool for embedding and verifying hidden information in PDF files.",
	Long: `Defender is a CLI tool that allows you to embed encrypted identification
information into PDF files without disrupting the reading experience, and
to extract and verify this information later.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Embed encrypted message into a PDF file.",
	Long: `The sign command reads a source PDF file, encrypts a user-specified
string (e.g., employee ID), and embeds it at the end of the file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if filePath == "" || message == "" || key == "" {
			return fmt.Errorf("all flags --file, --msg, and --key are required")
		}
		fmt.Printf("Attempting to sign file: %s with message: %s\n", filePath, message)
		err := injector.Sign(filePath, message, key)
		if err != nil {
			return fmt.Errorf("failed to sign file: %w", err)
		}
		fmt.Printf("Successfully signed file: %s_signed.pdf\n", filePath)
		return nil
	},
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Extract and verify hidden message from a PDF file.",
	Long: `The verify command attempts to read hidden data from a PDF, decrypt it,
and verify its integrity.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if filePath == "" || key == "" {
			return fmt.Errorf("both flags --file and --key are required")
		}
		fmt.Printf("Attempting to verify file: %s\n", filePath)
		extractedMsg, err := injector.Verify(filePath, key)
		if err != nil {
			return fmt.Errorf("failed to verify file: %w", err)
		}
		fmt.Printf("Verification successful. Extracted message: \"%s\"\n", extractedMsg)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(verifyCmd)

	signCmd.Flags().StringVarP(&filePath, "file", "f", "", "Source PDF file path")
	signCmd.Flags().StringVarP(&message, "msg", "m", "", "Message to embed (e.g., EmployeeID:123)")
	signCmd.Flags().StringVarP(&key, "key", "k", "", "Encryption key")

	verifyCmd.Flags().StringVarP(&filePath, "file", "f", "", "Target PDF file path")
	verifyCmd.Flags().StringVarP(&key, "key", "k", "", "Decryption key")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.SetOutput(os.Stderr)
		log.Fatalf("Error: %v", err)
	}
}

func main() {
	Execute()
}
