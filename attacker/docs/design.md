# Attacker Module Design Document

## 1. Overview
The **Attacker** module is a CLI tool designed to detect and remove hidden data (steganography) from PDF files. It operates on the principle that standard PDF readers ignore data appended after the End-Of-File (`%%EOF`) marker, a common technique used for simple steganography.

## 2. Architecture

### 2.1 Component Structure
```
attacker/
├── main.go           # Entry point, handles CLI argument parsing (Cobra/Flags)
├── core/             # Core business logic
│   ├── scanner.go    # Logic for detecting anomalies
│   └── cleaner.go    # Logic for sanitizing files
└── utils/            # Shared utilities (file I/O, byte manipulation)
```

### 2.2 Dependencies
- Standard Library: `os`, `io`, `bytes`, `fmt`, `flag` (or `spf13/cobra` if preferred for better CLI experience, but standard `flag` is sufficient for MVP).
- **Decision**: Use standard `flag` for zero-dependency simplicity unless complexity grows.

## 3. Detailed Design

### 3.1 Command: `scan`
**Goal**: Detect if a PDF file contains suspicious data after the EOF marker.

**Algorithm**:
1.  Open the target PDF file.
2.  Read the file from the end (reverse search) to find the last occurrence of the `%%EOF` marker.
    *   *Note*: PDF files can have multiple `%%EOF` markers (e.g., incremental updates). We are interested in the *physical* end of the valid PDF structure. However, simple append-steganography usually adds data *after* the final `%%EOF`.
    *   *Strategy*: Find the *last* valid `%%EOF` that conforms to PDF structure, or simply scan for the last occurrence of the byte sequence `%%EOF`.
    *   *Refinement*: To be robust, we should look for the last `%%EOF`. Any bytes following it (excluding optional whitespace `0x0D`, `0x0A`) are considered suspicious.
3.  Calculate `SuspiciousBytes = TotalFileSize - (EOF_Offset + Length("%%EOF"))`.
4.  Report status:
    *   **Clean**: `SuspiciousBytes <= Threshold` (allow small whitespace tolerance).
    *   **Suspicious**: `SuspiciousBytes > Threshold`.

### 3.2 Command: `clean`
**Goal**: Remove the suspicious data to sanitize the file.

**Algorithm**:
1.  Perform the same `scan` logic to locate the last `%%EOF`.
2.  Determine the `CutoffPoint` = `EOF_Offset + Length("%%EOF")`.
    *   *Option*: Preserve a trailing newline if present to be nice.
3.  Create a new file `{filename}_cleaned.pdf`.
4.  Copy bytes `0` to `CutoffPoint` from source to destination.
5.  Close files.

## 4. Implementation Plan

1.  **Setup**: Initialize `attacker` module structure.
2.  **Core Logic**: Implement `FindLastEOF` function in `core` package.
3.  **Scanner**: Implement `ScanPDF` using `FindLastEOF`.
4.  **Cleaner**: Implement `CleanPDF` using `FindLastEOF` and file truncation/copying.
5.  **CLI**: Wire up `main.go` to dispatch `scan` and `clean` commands.
6.  **Testing**: Create dummy PDFs with appended data to verify detection and cleaning.

## 5. Future Improvements (Anti-Forensics)
- **Metadata Scrubbing**: Remove metadata that might hint at the tool used or the original author.
- **Stream Analysis**: Deep scan of PDF streams for hidden data inside zlib-compressed blocks (Phase 2).
