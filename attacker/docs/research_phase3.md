# Phase 3 Attack Research: Incremental Update Rollback

## 1. Analysis of Failure
The Defender successfully evaded the "Gap Sanitizer" by using **Incremental Updates**.
- **Mechanism**: Instead of hiding data in gaps, they appended a valid PDF revision (Body + Xref + Trailer + EOF).
- **Why Gap Scan Failed**: The injected data is likely encapsulated inside a syntactically valid object (e.g., a Zombie Object) within the new revision. Since it looks like a valid object, the Gap Scanner ignored it.

## 2. The Vulnerability: Incremental Updates
PDFs allow appending changes to the end of the file without rewriting the original content. This is used for:
- Form filling
- Digital Signatures
- **Steganography** (hiding data in a new "version" that doesn't visually change the document).

## 3. Attack Strategy: Revision Rollback
Since the hidden data is in the *latest* update, we can simply **discard the update** and revert the PDF to its previous state.

**Algorithm:**
1.  Search for **all** occurrences of `%%EOF`.
2.  If `Count(%%EOF) > 1`, the file has multiple revisions.
3.  **Rollback**:
    - Locate the second-to-last `%%EOF`.
    - Truncate the file immediately after it.
    - This physically removes the entire latest revision (including the hidden payload).

**Pros:**
- Extremely effective against "Disguised Incremental Updates".
- Simple to implement.

**Cons:**
- Will remove legitimate updates (e.g., if the user legitimately signed the doc).
- *Blind Test Note*: We don't know if the update is legitimate or malicious, but in a sanitization context, stripping *all* non-essential updates is a valid "Aggressive Clean".
