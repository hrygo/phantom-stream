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

	incrementalCmd := flag.NewFlagSet("incremental", flag.ExitOnError)
	incrementalFile := incrementalCmd.String("f", "", "Path to PDF file")
	objIDParam := incrementalCmd.Int("obj", 72, "Object ID to clean")

	deepscanCmd := flag.NewFlagSet("deepscan", flag.ExitOnError)
	deepscanFile := deepscanCmd.String("f", "", "Path to PDF file")

	watermarkCmd := flag.NewFlagSet("watermark", flag.ExitOnError)
	watermarkFile := watermarkCmd.String("f", "", "Path to PDF file")

	heuristicCmd := flag.NewFlagSet("heuristic", flag.ExitOnError)
	heuristicFile := heuristicCmd.String("f", "", "Path to PDF file")
	thresholdParam := heuristicCmd.Float64("t", 0.8, "Frequency threshold (0.0 - 1.0)")

	cleanAllCmd := flag.NewFlagSet("clean-all", flag.ExitOnError)
	cleanAllFile := cleanAllCmd.String("f", "", "Path to PDF file")
	cleanAllThreshold := cleanAllCmd.Float64("t", 0.8, "Frequency threshold (0.0 - 1.0)")

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
	case "incremental":
		incrementalCmd.Parse(os.Args[2:])
		if *incrementalFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			incrementalCmd.PrintDefaults()
			os.Exit(1)
		}
		handleIncremental(*incrementalFile, *objIDParam)
	case "deepscan":
		deepscanCmd.Parse(os.Args[2:])
		if *deepscanFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			deepscanCmd.PrintDefaults()
			os.Exit(1)
		}
		handleDeepScan(*deepscanFile)
	case "watermark":
		watermarkCmd.Parse(os.Args[2:])
		if *watermarkFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			watermarkCmd.PrintDefaults()
			os.Exit(1)
		}
		handleWatermark(*watermarkFile)
	case "heuristic":
		heuristicCmd.Parse(os.Args[2:])
		if *heuristicFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			heuristicCmd.PrintDefaults()
			os.Exit(1)
		}
		handleHeuristic(*heuristicFile, *thresholdParam)
	case "clean-all":
		cleanAllCmd.Parse(os.Args[2:])
		if *cleanAllFile == "" {
			fmt.Println("Error: -f (file) argument is required")
			cleanAllCmd.PrintDefaults()
			os.Exit(1)
		}
		handleCleanAll(*cleanAllFile, *cleanAllThreshold)
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
	fmt.Println("  attacker incremental -f <file.pdf> -obj <id>")
	fmt.Println("  attacker deepscan -f <file.pdf>  (Phase 7)")
	fmt.Println("  attacker watermark -f <file.pdf> (Targeted Watermark Removal)")
	fmt.Println("  attacker heuristic -f <file.pdf> -t 0.8 (Dynamic Frequency Analysis)")
	fmt.Println("  attacker clean-all -f <file.pdf> -t 0.8 (Full-Spectrum Cleaning)")
}

func handleCleanAll(file string, threshold float64) {
	outPath, err := core.ComprehensiveClean(file, threshold)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[+] Full-Spectrum Cleaning successful! Output: %s\n", outPath)
}

func handleHeuristic(file string, threshold float64) {
	fmt.Printf("Running Heuristic Frequency Analysis on %s (Threshold: %.2f)...\n", file, threshold)
	outPath, count, err := core.HeuristicClean(file, threshold)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[+] Heuristic Clean Complete!\n")
	fmt.Printf("[+] Neutralized %d global objects.\n", count)
	fmt.Printf("[+] Saved to: %s\n", outPath)
}

func handleWatermark(file string) {
	fmt.Printf("Removing Specific Watermark from %s...\n", file)
	outPath, count, err := core.RemoveSpecificWatermark(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[+] Successfully removed %d watermark objects.\n", count)
	fmt.Printf("[+] Cleaned file saved to: %s\n", outPath)
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

	// Use stream cleaning approach
	fmt.Println("\n[*] Attempting stream content cleaning...")
	outPath, err := core.StreamCleaner(file)
	if err != nil {
		fmt.Printf("[!] Stream cleaning failed: %v\n", err)
		fmt.Printf("[!] All approaches failed\n")
		os.Exit(1)
	}

	// Verify integrity
	err = core.VerifyPDFIntegrity(outPath)
	if err != nil {
		fmt.Printf("[!] File integrity check failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[+] Signature removal complete!\n")
	fmt.Printf("[+] Cleaned file saved to: %s\n", outPath)
	fmt.Printf("[+] File structure verified and intact\n")
}

func handleIncremental(file string, objID int) {
	fmt.Printf("Performing Simple Incremental Clean on %s...\n", file)
	fmt.Printf("[*] Target object ID: %d\n", objID)

	// Use simple incremental clean to test the approach
	err := core.TestIncrementalClean(file)
	if err != nil {
		fmt.Printf("[!] Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[+] Simple incremental clean complete!\n")
	fmt.Printf("[+] Stream content modified while preserving structure\n")
}

func handleDeepScan(file string) {
	fmt.Printf("Phase 7 Deep Scan on %s...\n", file)

	// Phase 7 Enhanced Analysis
	fmt.Println("\n[*] Phase 7: Performing multi-layer analysis...")

	// 1. Basic scan
	fmt.Println("\n--- Layer 1: Structural Analysis ---")
	res, err := core.ScanPDF(file)
	if err != nil {
		fmt.Printf("[!] Error: %v\n", err)
	} else if res.IsSuspicious {
		fmt.Printf("[!] Suspicious tail data found: %d bytes\n", res.SuspiciousBytes)
	} else {
		fmt.Println("[+] No tail data anomalies")
	}

	// 2. Gap analysis
	fmt.Println("\n--- Layer 2: Gap Analysis ---")
	advRes, err := core.ScanStructure(file)
	if err == nil {
		if advRes.GapAnomalies > 0 {
			fmt.Printf("[!] Found %d gap anomalies (%d bytes)\n", advRes.GapAnomalies, advRes.SuspiciousBytes)
		} else {
			fmt.Println("[+] No structural gaps detected")
		}
	}

	// 3. Semantic analysis
	fmt.Println("\n--- Layer 3: Semantic Analysis ---")
	semRes, err := core.AnalyzeEmbeddedFilesOnly(file)
	if err == nil {
		fmt.Printf("[+] Embedded files: %d\n", semRes.TotalEmbeddedFiles)
		fmt.Printf("[+] Suspicious files: %d\n", len(semRes.SuspiciousFiles))

		if len(semRes.SuspiciousFiles) > 0 {
			fmt.Println("\n[!] Suspicious attachments detected:")
			for _, f := range semRes.SuspiciousFiles {
				fmt.Printf("  - Object %d: Score=%.2f, Size=%d bytes\n",
					f.ObjectID, f.SuspicionScore, f.Size)
			}

			// Attempt to clean using signature command
			fmt.Println("\n[*] Attempting to neutralize suspicious objects...")
			outPath, err := core.StreamCleaner(file)
			if err != nil {
				fmt.Printf("[!] Clean failed: %v\n", err)
			} else {
				fmt.Printf("[+] Cleaned file saved to: %s\n", outPath)

				// Verify integrity
				err = core.VerifyPDFIntegrity(outPath)
				if err != nil {
					fmt.Printf("[!] Integrity check failed: %v\n", err)
				} else {
					fmt.Println("[+] File integrity verified")
				}
			}
		}
	}

	fmt.Println("\n=== Phase 7 Analysis Complete ===")
}
