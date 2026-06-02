package documents

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const maxDoctreeCollectionName = 120

var nonNameSegment = regexp.MustCompile(`[^a-z0-9]+`)

// DoctreeCollectionName returns the deterministic dt_* collection name for a leaf folder.
// leafRel is relative to the doctree stub ("." for markdown directly under the stub).
func DoctreeCollectionName(stub, leafRel string) string {
	leafRel = filepath.ToSlash(leafRel)
	parts := []string{"dt", sanitizeNameSegment(stub)}
	if leafRel != "." && leafRel != "" {
		for _, seg := range strings.Split(leafRel, "/") {
			if seg == "" || seg == "." {
				continue
			}
			parts = append(parts, sanitizeNameSegment(seg))
		}
	}
	name := collapseUnderscores(strings.Join(parts, "_"))
	if len(name) <= maxDoctreeCollectionName {
		return name
	}
	key := stub + "/" + leafRel
	sum := sha256.Sum256([]byte(key))
	hash := fmt.Sprintf("%x", sum)[:6]
	keep := maxDoctreeCollectionName - 1 - len(hash)
	if keep < 8 {
		keep = 8
	}
	return name[:keep] + "_" + hash
}

// LeafRelFromDestRoot returns the leaf path relative to stub for a binding destRoot.
func LeafRelFromDestRoot(stub, destRoot string) string {
	destRoot = filepath.ToSlash(destRoot)
	stub = filepath.ToSlash(stub)
	if destRoot == stub {
		return "."
	}
	return strings.TrimPrefix(destRoot, stub+"/")
}

func sanitizeNameSegment(seg string) string {
	seg = strings.ToLower(seg)
	return nonNameSegment.ReplaceAllString(seg, "_")
}

func collapseUnderscores(s string) string {
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	return strings.Trim(s, "_")
}
