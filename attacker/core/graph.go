package core

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

// ObjectNode represents a PDF object
type ObjectNode struct {
	ID         int
	Gen        int
	Offset     int64
	Length     int64
	References []int // IDs of objects this object points to
}

// GraphResult holds the analysis of the PDF object graph
type GraphResult struct {
	TotalObjects     int
	ReachableObjects int
	ZombieObjects    []int // IDs of unreferenced objects
	ObjectMap        map[int]*ObjectNode
}

// AnalyzeObjectGraph builds a map of objects and references to find zombies.
func AnalyzeObjectGraph(filePath string) (*GraphResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 1. Find all object definitions: "ID Gen obj"
	// We use a map to store them: ID -> Node
	objects := make(map[int]*ObjectNode)

	objRegex := regexp.MustCompile(`(\d+)\s+(\d+)\s+obj`)
	endObjRegex := regexp.MustCompile(`endobj`)

	objMatches := objRegex.FindAllSubmatchIndex(content, -1)

	for _, match := range objMatches {
		// match[0]-match[1] is the full match
		// match[2]-match[3] is ID
		// match[4]-match[5] is Gen

		idStr := string(content[match[2]:match[3]])
		genStr := string(content[match[4]:match[5]])

		id, _ := strconv.Atoi(idStr)
		gen, _ := strconv.Atoi(genStr)

		startOffset := match[0]

		// Find where this object ends
		// Search for 'endobj' starting from startOffset
		endLoc := endObjRegex.FindIndex(content[startOffset:])
		if endLoc == nil {
			continue // Broken object?
		}

		length := int64(endLoc[1]) // Relative to startOffset

		// Extract content of the object to find references
		objContent := content[startOffset : startOffset+endLoc[1]]
		refs := findReferences(objContent)

		objects[id] = &ObjectNode{
			ID:         id,
			Gen:        gen,
			Offset:     int64(startOffset),
			Length:     length,
			References: refs,
		}
	}

	// 2. Find Root (Trailer)
	// The trailer contains "/Root X Y R"
	trailerRegex := regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	rootMatch := trailerRegex.FindSubmatch(content)

	if rootMatch == nil {
		return nil, fmt.Errorf("could not find /Root in trailer")
	}

	rootID, _ := strconv.Atoi(string(rootMatch[1]))

	// 3. Traverse Graph (BFS)
	visited := make(map[int]bool)
	queue := []int{rootID}
	visited[rootID] = true

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		node, exists := objects[currentID]
		if !exists {
			continue
		}

		for _, refID := range node.References {
			if !visited[refID] {
				visited[refID] = true
				queue = append(queue, refID)
			}
		}
	}

	// 4. Identify Zombies
	var zombies []int
	for id := range objects {
		if !visited[id] {
			zombies = append(zombies, id)
		}
	}

	return &GraphResult{
		TotalObjects:     len(objects),
		ReachableObjects: len(visited),
		ZombieObjects:    zombies,
		ObjectMap:        objects,
	}, nil
}

func findReferences(data []byte) []int {
	// Pattern: "ID Gen R"
	// Note: This is a heuristic. It might match data inside streams.
	refRegex := regexp.MustCompile(`(\d+)\s+(\d+)\s+R`)
	matches := refRegex.FindAllSubmatch(data, -1)

	var refs []int
	for _, m := range matches {
		id, _ := strconv.Atoi(string(m[1]))
		refs = append(refs, id)
	}
	return refs
}
