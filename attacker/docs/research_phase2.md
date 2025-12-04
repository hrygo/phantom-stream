# Phase 2 Attack Research: Advanced PDF Steganography Detection

## 1. Threat Model Update
The Defender has acknowledged that "EOF Append" (Phase 1) is compromised. They will likely move to **Phase 2: Internal Storage**.

**Likely Defender Tactics:**
1.  **Unreferenced Objects**: Adding valid PDF objects (e.g., `100 0 obj ... endobj`) that are not linked from the `Catalog` or `xref` table.
2.  **Stream Injection**: Hiding data inside valid streams (e.g., Image or Content streams), possibly after the valid data but before `endstream`.
3.  **Whitespace/Comment Injection**: Hiding data in comments (`%`) or excessive whitespace between objects.
4.  **Incremental Updates**: Abusing the PDF revision history feature to hide data in a "previous version" that is never shown.

## 2. Attack Strategy Upgrade

To counter these, the Attacker tool needs to evolve from a simple "File Tail Scanner" to a **Structural PDF Analyzer**.

### 2.1 New Scanner Capabilities
- **Cross-Reference Analysis**: Parse the `xref` table to know exactly where every valid object *should* be.
- **Gap Analysis**: Scan the physical file gaps between defined objects. Any non-whitespace bytes between `endobj` and the next `obj` (or `xref`) are suspicious.
- **Object Traversal**: Walk the object tree starting from `Root`. Mark all reachable objects. Any object in the file that is *not* marked is "Unreferenced" (Zombie Object) and likely contains hidden data.

### 2.2 New Cleaner Capabilities
- **Garbage Collection (GC)**: Instead of just truncating the file, we rewrite the PDF from scratch.
    - Read valid objects.
    - Write them to a new file sequentially.
    - Generate a fresh `xref` table.
    - This naturally drops *all* unreferenced objects, hidden comments, and inter-object gaps.

## 3. Implementation Plan (Hunter v2)

1.  **PDF Parser Integration**: We need a way to parse PDF structure.
    - *Option A*: Write a minimal parser (Complex, high effort).
    - *Option B*: Use a library like `rsc.io/pdf` or `github.com/pdfcpu/pdfcpu` (High reliability).
    - *Decision*: For a "Hunter" tool, we want granular control. We will implement a **Gap Scanner** first (low effort, high ROI for "hidden between objects" attacks).

2.  **Gap Scanner Logic**:
    - Regex/Byte-search for `obj` and `endobj` markers.
    - Record all `[start, end]` intervals of valid objects.
    - Check bytes *between* intervals.

3.  **Rewriter (The Ultimate Cleaner)**:
    - If we can parse the objects, we can just copy the *body* of valid objects to a new file and rebuild the index. This sanitizes *everything*.

## 4. Immediate Action
- Implement a **"Gap Scanner"** in `attacker/core/scanner_v2.go`.
- This will detect if Defender tries to hide data *between* objects or in comments.
