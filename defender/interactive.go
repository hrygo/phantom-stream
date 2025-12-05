package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"defender/injector"
)

// ANSI Colors
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorBold   = "\033[1m"
)

func runInteractive() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		clearScreen()
		fmt.Println(ColorBlue + "==================================================" + ColorReset)
		fmt.Println(ColorBold + "   🛡️  PhantomGuard - PDF Protection Tool" + ColorReset)
		fmt.Println(ColorBlue + "==================================================" + ColorReset)
		fmt.Println("")
		fmt.Println("1. " + ColorBold + "🔒 Protect PDF" + ColorReset + "   (Embed invisible watermark)")
		fmt.Println("2. " + ColorBold + "🔍 Verify PDF" + ColorReset + "    (Extract & verify watermark)")
		fmt.Println("3. " + ColorRed + "🚪 Exit" + ColorReset)
		fmt.Println("")
		fmt.Print(ColorCyan + "Select option (1-3): " + ColorReset)

		if !scanner.Scan() {
			return
		}
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			interactiveProtect(scanner)
		case "2":
			interactiveVerify(scanner)
		case "3":
			fmt.Println("\n" + ColorBlue + "Stay safe! 👋" + ColorReset)
			os.Exit(0)
		default:
			fmt.Println(ColorRed + "\nInvalid option. Please try again." + ColorReset)
			time.Sleep(1 * time.Second)
		}
	}
}

func interactiveProtect(scanner *bufio.Scanner) {
	fmt.Println("\n" + ColorYellow + "--- [PROTECT] Protect PDF Mode ---" + ColorReset)

	// Step 1: File
	var path string
	for {
		fmt.Print("\n" + ColorBold + "[Step 1/4] Enter PDF file path" + ColorReset + " (drag & drop file here):\n> ")
		if !scanner.Scan() {
			return
		}
		path = cleanPath(scanner.Text())
		if path == "" {
			continue
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Println(ColorRed + "[ERROR] File does not exist. Please try again." + ColorReset)
			continue
		}
		break
	}

	// Step 2: Message
	fmt.Print("\n" + ColorBold + "[Step 2/4] Enter watermark message" + ColorReset + " (e.g. 'Confidential', 'User:Alice'):\n> ")
	if !scanner.Scan() {
		return
	}
	msg := strings.TrimSpace(scanner.Text())
	if msg == "" {
		msg = "Protected Document"
		fmt.Println(ColorYellow + "[*] Using default message: 'Protected Document'" + ColorReset)
	}

	// Step 3: Key
	fmt.Print("\n" + ColorBold + "[Step 3/4] Enter encryption key" + ColorReset + " (32 chars) [Press Enter to auto-generate]:\n> ")
	if !scanner.Scan() {
		return
	}
	key := strings.TrimSpace(scanner.Text())

	if key == "" {
		k := make([]byte, 16)
		if _, err := rand.Read(k); err != nil {
			fmt.Printf(ColorRed+"Error generating key: %v\n"+ColorReset, err)
			return
		}
		key = hex.EncodeToString(k)
		fmt.Printf(ColorGreen+"[*] Generated Key: %s\n"+ColorReset, key)
	}

	if len(key) != 32 {
		fmt.Println(ColorRed + "[ERROR] Key must be exactly 32 characters." + ColorReset)
		waitForEnter(scanner)
		return
	}

	// Step 4: Protection Level
	fmt.Println("\n" + ColorBold + "[Step 4/4] Select Protection Level:" + ColorReset)
	fmt.Println("1. " + ColorGreen + "Standard" + ColorReset + " (Attachment only) - Compatible, easy to detect")
	fmt.Println("2. " + ColorYellow + "Stealth" + ColorReset + "  (Attachment + SMask) - Harder to detect")
	fmt.Println("3. " + ColorRed + "Maximum" + ColorReset + "  (All Anchors: Attachment + SMask + Content + Visual) - Max resilience")
	fmt.Println("4. " + ColorBlue + "Custom" + ColorReset + "   (Select specific anchors)")
	fmt.Print("> ")

	if !scanner.Scan() {
		return
	}
	level := strings.TrimSpace(scanner.Text())

	var selectedAnchors []string

	switch level {
	case "1":
		selectedAnchors = []string{"Attachment"}
	case "2":
		selectedAnchors = []string{"Attachment", "SMask"}
	case "3":
		selectedAnchors = nil // All
	case "4":
		// Custom selection
		fmt.Println("\nAvailable Anchors: Attachment, SMask, Content, Visual")
		fmt.Print("Enter anchor names separated by comma (e.g. 'Attachment,Visual'):\n> ")
		if !scanner.Scan() {
			return
		}
		input := scanner.Text()
		parts := strings.Split(input, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				selectedAnchors = append(selectedAnchors, p)
			}
		}
	default:
		fmt.Println(ColorYellow + "[*] Defaulting to Maximum protection" + ColorReset)
		selectedAnchors = nil
	}

	fmt.Println("\n" + ColorBlue + "[*] Processing..." + ColorReset)

	// Execute
	err := injector.Sign(path, msg, key, "", selectedAnchors)
	if err != nil {
		fmt.Printf(ColorRed+"[ERROR] Protection failed: %v\n"+ColorReset, err)
	} else {
		// Calculate output path to show user
		dir := filepath.Dir(path)
		base := filepath.Base(path)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)
		outPath := filepath.Join(dir, name+"_signed"+ext)

		fmt.Println("\n" + ColorGreen + "[SUCCESS] File protected." + ColorReset)
		fmt.Printf("[FILE] Output File: %s\n", outPath)
		fmt.Println(ColorYellow + "--------------------------------------------------" + ColorReset)
		fmt.Printf("[KEY] Key: "+ColorBold+"%s"+ColorReset+"\n", key)
		fmt.Println(ColorYellow + "[WARNING] IMPORTANT: Save this key! It is required for verification." + ColorReset)
		fmt.Println(ColorYellow + "--------------------------------------------------" + ColorReset)
	}

	waitForEnter(scanner)
}

func interactiveVerify(scanner *bufio.Scanner) {
	fmt.Println("\n" + ColorYellow + "--- [VERIFY] Verify PDF Mode ---" + ColorReset)

	var path string
	for {
		fmt.Print("\n" + ColorBold + "[Step 1/3] Enter PDF file path:\n> " + ColorReset)
		if !scanner.Scan() {
			return
		}
		path = cleanPath(scanner.Text())
		if path == "" {
			continue
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Println(ColorRed + "[ERROR] File does not exist. Please try again." + ColorReset)
			continue
		}
		break
	}

	fmt.Print("\n" + ColorBold + "[Step 2/3] Enter decryption key:\n> " + ColorReset)
	if !scanner.Scan() {
		return
	}
	key := strings.TrimSpace(scanner.Text())

	// Step 3: Verification Mode
	fmt.Println("\n" + ColorBold + "[Step 3/3] Select Verification Mode:" + ColorReset)
	fmt.Println("1. " + ColorGreen + "Auto" + ColorReset + " (Try all, stop at first success)")
	fmt.Println("2. " + ColorBlue + "Custom" + ColorReset + " (Verify specific anchors)")
	fmt.Print("> ")

	if !scanner.Scan() {
		return
	}
	mode := strings.TrimSpace(scanner.Text())

	var selectedAnchors []string

	if mode == "2" {
		fmt.Println("\nAvailable Anchors: Attachment, SMask, Content")
		fmt.Println(ColorYellow + "Note: Visual watermark does not support extraction/verification" + ColorReset)
		fmt.Print("Enter anchor names separated by comma (e.g. 'SMask'):\n> ")
		if !scanner.Scan() {
			return
		}
		input := scanner.Text()
		parts := strings.Split(input, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				selectedAnchors = append(selectedAnchors, p)
			}
		}
	}

	fmt.Println("\n" + ColorBlue + "[*] Verifying..." + ColorReset)
	msg, anchorName, err := injector.Verify(path, key, selectedAnchors)
	if err != nil {
		fmt.Printf(ColorRed+"[ERROR] Verification Failed: %v\n"+ColorReset, err)
		fmt.Println(ColorYellow + "Possible reasons: Wrong key, file tampered, or not protected." + ColorReset)
	} else {
		fmt.Println("\n" + ColorGreen + "[SUCCESS] Verification Successful!" + ColorReset)
		fmt.Printf("Found via: "+ColorBold+"%s"+ColorReset+"\n", anchorName)
		fmt.Printf("Hidden Message: "+ColorBold+"%s"+ColorReset+"\n", msg)
	}

	waitForEnter(scanner)
}

func cleanPath(p string) string {
	return strings.Trim(strings.TrimSpace(p), "\"'")
}

func waitForEnter(scanner *bufio.Scanner) {
	fmt.Println("\nPress Enter to return to menu...")
	scanner.Scan()
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
