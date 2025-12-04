package core

import (
	"fmt"
	"os"
)

// AnalyzeEmbeddedFilesOnly analyzes embedded files without removing them
func AnalyzeEmbeddedFilesOnly(filePath string) (*SemanticResult, error) {
	return AnalyzeEmbeddedFiles(filePath)
}

// ReportSuspiciousAttachments creates a detailed report of suspicious attachments
func ReportSuspiciousAttachments(filePath string) (string, *SemanticResult, error) {
	result, err := AnalyzeEmbeddedFiles(filePath)
	if err != nil {
		return "", nil, err
	}

	report := fmt.Sprintf("# Suspicious Attachments Analysis Report\n\n")
	report += fmt.Sprintf("File: %s\n", filePath)
	report += fmt.Sprintf("Total embedded files: %d\n", result.TotalEmbeddedFiles)
	report += fmt.Sprintf("Suspicious files: %d\n", len(result.SuspiciousFiles))
	report += fmt.Sprintf("Clean files: %d\n\n", len(result.CleanFiles))

	if len(result.SuspiciousFiles) > 0 {
		report += "## Suspicious Files Detected:\n\n"
		for i, file := range result.SuspiciousFiles {
			report += fmt.Sprintf("### Suspicious File #%d\n", i+1)
			report += fmt.Sprintf("- Object ID: %d\n", file.ObjectID)
			report += fmt.Sprintf("- Size: %d bytes\n", file.Size)
			report += fmt.Sprintf("- Suspicion Score: %.2f\n", file.SuspicionScore)
			report += fmt.Sprintf("- Is Compressed: %v\n", file.IsCompressed)
			report += fmt.Sprintf("- Filter: %s\n", file.Filter)
			report += fmt.Sprintf("- Content Entropy: %.2f\n", file.ContentEntropy)
			report += "- Reasons:\n"
			for _, reason := range file.Reasons {
				report += fmt.Sprintf("  * %s\n", reason)
			}
			report += "\n"
		}
		report += "## Recommendation\n\n"
		report += "The following suspicious attachments have been detected. Manual review is recommended before removal.\n"
	} else {
		report += "## No Suspicious Attachments Detected\n\n"
		report += "The file appears to be clean based on semantic analysis.\n"
	}

	// Write report to file
	reportPath := filePath + "_semantic_report.txt"
	err = os.WriteFile(reportPath, []byte(report), 0644)
	if err != nil {
		return "", nil, err
	}

	return reportPath, result, nil
}