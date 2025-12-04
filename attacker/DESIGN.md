# Attacker Module Design Document

## 1. Overview
The **Attacker** module is a CLI tool designed to detect and remove hidden data (steganography) from PDF files. It operates on the principle that standard PDF readers ignore data appended after the End-Of-File (`%%EOF`) marker.

## 2. Architecture

### 2.1 Component Structure
```
attacker/
├── main.go           # Entry point
├── core/             # Core business logic
│   ├── scanner.go    # Logic for detecting anomalies
│   └── cleaner.go    # Logic for sanitizing files
└── test_data/        # Test files (Isolated)
```

## 3. Detailed Design

### 3.1 Command: `scan`
**Goal**: Detect if a PDF file contains suspicious data after the EOF marker.
**Algorithm**:
1.  Read file.
2.  Find last `%%EOF`.
3.  Calculate `SuspiciousBytes = TotalFileSize - (EOF_Offset + 5)`.
4.  If `SuspiciousBytes > 0` (ignoring whitespace), report **SUSPICIOUS**.

### 3.2 Command: `clean`
**Goal**: Remove suspicious data.
**Algorithm**:
1.  Find last `%%EOF`.
2.  Truncate file at `EOF_Offset + 5` (preserving optional newline).
3.  Save as `{filename}_cleaned.pdf`.
