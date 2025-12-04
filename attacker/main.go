package main

import (
	"flag"
	"fmt"
	"os"

	"attacker/core"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	scanCmd := flag.NewFlagSet("scan", flag.ExitOnError)
	scanFile := scanCmd.String("f", "", "Path to PDF file")

	cleanCmd := flag.NewFlagSet("clean", flag.ExitOnError)
	cleanFile := cleanCmd.String("f", "", "Path to PDF file")

	switch os.Args[1] {
	case "scan":
		scanCmd.Parse(os.Args[2:])
		if *scanFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			scanCmd.PrintDefaults()
			os.Exit(1)
		}
		handleScan(*scanFile)
	case "clean":
		cleanCmd.Parse(os.Args[2:])
		if *cleanFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			cleanCmd.PrintDefaults()
			os.Exit(1)
		}
		handleClean(*cleanFile)
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Attacker CLI Tool")
	fmt.Println("Usage:")
	fmt.Println("  attacker scan -f <file.pdf>")
	fmt.Println("  attacker clean -f <file.pdf>")
}

func handleScan(file string) {
	fmt.Printf("Scanning %s...\n", file)
	res, err := core.ScanPDF(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if res.IsSuspicious {
		fmt.Println("[!] Status: SUSPICIOUS (Tail Data)")
		fmt.Printf("[!] Suspicious Data: %d bytes found after EOF\n", res.SuspiciousBytes)
	} else {
		// Run Advanced Scan if tail is clean
		advRes, err := core.ScanStructure(file)
		if err == nil && advRes.GapAnomalies > 0 {
			fmt.Println("[!] Status: SUSPICIOUS (Internal Gaps)")
			fmt.Printf("[!] Found %d anomalies between objects (%d bytes)\n", advRes.GapAnomalies, advRes.SuspiciousBytes)
		} else {
			fmt.Println("[+] Status: CLEAN")
			fmt.Println("[+] No suspicious data found.")
		}
	}
}

func handleClean(file string) {
	fmt.Printf("Cleaning %s...\n", file)
	outPath, removed, err := core.CleanPDF(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[+] Cleaned file saved to: %s\n", outPath)
	fmt.Printf("[+] Removed %d bytes of data.\n", removed)
}
