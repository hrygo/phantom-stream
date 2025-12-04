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

	sanitizeCmd := flag.NewFlagSet("sanitize", flag.ExitOnError)
	sanitizeFile := sanitizeCmd.String("f", "", "Path to PDF file")

	rollbackCmd := flag.NewFlagSet("rollback", flag.ExitOnError)
	rollbackFile := rollbackCmd.String("f", "", "Path to PDF file")

	pruneCmd := flag.NewFlagSet("prune", flag.ExitOnError)
	pruneFile := pruneCmd.String("f", "", "Path to PDF file")

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
	case "sanitize":
		sanitizeCmd.Parse(os.Args[2:])
		if *sanitizeFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			sanitizeCmd.PrintDefaults()
			os.Exit(1)
		}
		handleSanitize(*sanitizeFile)
	case "rollback":
		rollbackCmd.Parse(os.Args[2:])
		if *rollbackFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			rollbackCmd.PrintDefaults()
			os.Exit(1)
		}
		handleRollback(*rollbackFile)
	case "prune":
		pruneCmd.Parse(os.Args[2:])
		if *pruneFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			pruneCmd.PrintDefaults()
			os.Exit(1)
		}
		handlePrune(*pruneFile)
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
	fmt.Println("  attacker sanitize -f <file.pdf>")
	fmt.Println("  attacker rollback -f <file.pdf>")
	fmt.Println("  attacker prune -f <file.pdf>")
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

func handleSanitize(file string) {
	fmt.Printf("Sanitizing (Gap Cleaning) %s...\n", file)
	outPath, sanitizedBytes, err := core.SanitizeGaps(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if sanitizedBytes == 0 {
		fmt.Println("[+] No suspicious gaps found. File is clean.")
	} else {
		fmt.Printf("[+] Sanitized file saved to: %s\n", outPath)
		fmt.Printf("[+] Overwrote %d bytes of suspicious data in gaps.\n", sanitizedBytes)
	}
}

func handleRollback(file string) {
	fmt.Printf("Attempting Rollback on %s...\n", file)
	outPath, res, err := core.RollbackPDF(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[+] Rollback successful!\n")
	fmt.Printf("[+] Found %d revisions.\n", res.RevisionsFound)
	fmt.Printf("[+] Reverted to previous version (Size: %d -> %d bytes).\n", res.OriginalSize, res.NewSize)
	fmt.Printf("[+] Saved to: %s\n", outPath)
}

func handlePrune(file string) {
	fmt.Printf("Pruning Zombie Objects from %s...\n", file)
	outPath, count, err := core.PruneZombies(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if count == 0 {
		fmt.Println("[+] No zombie objects found.")
	} else {
		fmt.Printf("[+] Pruned %d zombie objects.\n", count)
		fmt.Printf("[+] Saved to: %s\n", outPath)
	}
}
