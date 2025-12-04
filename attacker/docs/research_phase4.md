# Phase 4 Attack Research: Object Graph Analysis (The "Zombie" Hunter)

## 1. Threat Forecast
The Defender's next logical step is to hide data in **Valid but Unreferenced Objects** (Zombie Objects).
- **Technique**: Insert a valid PDF object (e.g., `999 0 obj <payload> endobj`) into the file body.
- **Evasion**:
    - It passes `Tail Scan` (it's before EOF).
    - It passes `Gap Scan` (it looks like a valid object, not a gap).
    - It passes `Rollback` (if they rewrite the file instead of appending).

## 2. Attack Strategy: Reachability Analysis (Garbage Collection)
To defeat this, we must treat the PDF as a **Directed Graph**.
- **Nodes**: PDF Objects (`obj`).
- **Edges**: References (`X Y R`).
- **Roots**: The `Trailer` dictionary points to the `Root` (Catalog).

**Algorithm (The "Pruner"):**
1.  **Identify all Objects**: Scan the file to find every `ID Gen obj` definition.
2.  **Identify all References**: Scan the file (or specific streams) for the pattern `ID Gen R`.
3.  **Build Graph**: Map which objects point to which.
4.  **Traverse**: Start from the `Trailer` -> `Root` and perform a BFS/DFS to find all *reachable* objects.
5.  **Detect Zombies**: Any object found in Step 1 that is *not* visited in Step 4 is a **Zombie**.
6.  **Action**: Report or Delete (Prune) these objects.

## 3. Implementation Challenges (Regex-based)
- **False Positives**: A string `(1 0 R)` inside a text stream might look like a reference but isn't.
- **False Negatives**: Obfuscated references (hex encoded) might be missed.
- **Mitigation**: For a "Blind Test" MVP, we will assume standard encoding. We will be aggressive: if an object is not clearly referenced, we flag it.

## 4. Plan
- Implement `core/graph.go`:
    - `BuildObjectGraph(content []byte)`
    - `FindUnreferencedObjects()`
- Add `prune` command to CLI.
