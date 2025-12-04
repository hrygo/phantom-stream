package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

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

	semanticCmd := flag.NewFlagSet("semantic", flag.ExitOnError)
	semanticFile := semanticCmd.String("f", "", "Path to PDF file")

	reportCmd := flag.NewFlagSet("report", flag.ExitOnError)
	reportFile := reportCmd.String("f", "", "Path to PDF file")

	detectCmd := flag.NewFlagSet("detect", flag.ExitOnError)
	detectFile := detectCmd.String("f", "", "Path to PDF file")

	signatureCmd := flag.NewFlagSet("signature", flag.ExitOnError)
	signatureFile := signatureCmd.String("f", "", "Path to PDF file")

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
	case "semantic":
		semanticCmd.Parse(os.Args[2:])
		if *semanticFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			semanticCmd.PrintDefaults()
			os.Exit(1)
		}
		handleSemantic(*semanticFile)
	case "report":
		reportCmd.Parse(os.Args[2:])
		if *reportFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			reportCmd.PrintDefaults()
			os.Exit(1)
		}
		handleReport(*reportFile)
	case "detect":
		detectCmd.Parse(os.Args[2:])
		if *detectFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			detectCmd.PrintDefaults()
			os.Exit(1)
		}
		handleDetect(*detectFile)
	case "signature":
		signatureCmd.Parse(os.Args[2:])
		if *signatureFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			signatureCmd.PrintDefaults()
			os.Exit(1)
		}
		handleSignature(*signatureFile)
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
	fmt.Println("  attacker semantic -f <file.pdf>")
	fmt.Println("  attacker report -f <file.pdf>")
	fmt.Println("  attacker detect -f <file.pdf>")
	fmt.Println("  attacker signature -f <file.pdf>")
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

func handleSemantic(file string) {
	fmt.Printf("Performing Semantic Analysis on %s...\n", file)
	outPath, result, err := core.RemoveSuspiciousAttachments(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[+] Semantic Analysis Complete!\n")
	fmt.Printf("[+] Total embedded files found: %d\n", result.TotalEmbeddedFiles)
	fmt.Printf("[+] Suspicious files detected: %d\n", len(result.SuspiciousFiles))
	fmt.Printf("[+] Clean files: %d\n", len(result.CleanFiles))

	if len(result.SuspiciousFiles) == 0 {
		fmt.Println("[+] No suspicious attachments found. File is clean.")
		return
	}

	fmt.Println("\n[!] Suspicious Attachments:")
	for _, file := range result.SuspiciousFiles {
		fmt.Printf("  - Object %d: Size=%d bytes, Score=%.2f\n", file.ObjectID, file.Size, file.SuspicionScore)
		fmt.Printf("    Reasons: %s\n", strings.Join(file.Reasons, "; "))
	}

	fmt.Printf("\n[+] Suspicious attachments removed!\n")
	fmt.Printf("[+] Total suspicious bytes removed: %d\n", result.TotalSuspiciousBytes)
	fmt.Printf("[+] Sanitized file saved to: %s\n", outPath)
}

func handleReport(file string) {
	fmt.Printf("Generating Suspicious Attachment Report for %s...\n", file)
	reportPath, result, err := core.ReportSuspiciousAttachments(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[+] Report Generation Complete!\n")
	fmt.Printf("[+] Total embedded files found: %d\n", result.TotalEmbeddedFiles)
	fmt.Printf("[+] Suspicious files detected: %d\n", len(result.SuspiciousFiles))
	fmt.Printf("[+] Clean files: %d\n", len(result.CleanFiles))
	fmt.Printf("[+] Detailed report saved to: %s\n", reportPath)

	if len(result.SuspiciousFiles) > 0 {
		fmt.Println("\n[!] Suspicious Attachments Found:")
		for _, file := range result.SuspiciousFiles {
			fmt.Printf("  - Object %d: Score=%.2f\n", file.ObjectID, file.SuspicionScore)
		}
		fmt.Println("\n[*] Use 'attacker semantic' to remove suspicious attachments")
	} else {
		fmt.Println("\n[+] No suspicious attachments detected.")
	}
}

func handleDetect(file string) {
	fmt.Printf("Performing Read-Only Detection on %s...\n", file)
	result, err := core.AnalyzeEmbeddedFilesOnly(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n[+] Detection Complete (No Modification Made)\n")
	fmt.Printf("[+] File: %s\n", file)
	fmt.Printf("[+] Total embedded files: %d\n", result.TotalEmbeddedFiles)
	fmt.Printf("[+] Suspicious files: %d\n", len(result.SuspiciousFiles))
	fmt.Printf("[+] Clean files: %d\n", len(result.CleanFiles))

	if len(result.SuspiciousFiles) > 0 {
		fmt.Println("\n[!] Threat Assessment:")
		for _, file := range result.SuspiciousFiles {
			fmt.Printf("\n  Suspicious Attachment Detected:\n")
			fmt.Printf("  - Object ID: %d\n", file.ObjectID)
			fmt.Printf("  - File Size: %d bytes\n", file.Size)
			fmt.Printf("  - Risk Score: %.2f/3.00\n", file.SuspicionScore)
			fmt.Printf("  - Compression: %v\n", file.IsCompressed)
			fmt.Printf("  - Risk Factors:\n")
			for _, reason := range file.Reasons {
				fmt.Printf("    * %s\n", reason)
			}
		}
		fmt.Println("\n[!] WARNING: This file contains suspicious embedded content")
		fmt.Println("[!] File has NOT been modified to preserve integrity")
	} else {
		fmt.Println("\n[+] No threats detected. File appears clean.")
	}
}

func handleSignature(file string) {
	fmt.Printf("Removing Signature/Tracking Data from %s...\n", file)

	// Try cleaning embedded files reference first (most conservative)
	fmt.Println("\n[*] Attempting catalog cleaning (removing EmbeddedFiles reference)...")
	outPath, err := core.CleanEmbeddedFilesReference(file)
	if err != nil {
		fmt.Printf("[!] Catalog cleaning failed: %v\n", err)
		fmt.Println("[*] Trying stream content cleaning...")

		// Try to clean the actual stream content
		outPath2, err := core.StreamCleaner(file)
		if err != nil {
			fmt.Printf("[!] Stream cleaning failed: %v\n", err)
			fmt.Printf("[!] All approaches failed\n")
			os.Exit(1)
		}

		outPath = outPath2
	}

	// Verify integrity
	err = core.VerifyFileIntegrity(outPath)
	if err != nil {
		fmt.Printf("[!] File integrity check failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[+] Signature removal complete!\n")
	fmt.Printf("[+] Cleaned file saved to: %s\n", outPath)
	fmt.Printf("[+] File structure verified and intact\n")
}
