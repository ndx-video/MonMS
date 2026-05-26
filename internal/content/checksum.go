package content

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

// ChecksumExport returns a stable sha256:hex checksum for editorial snapshots (PUB-08).
func ChecksumExport(payload any) (string, error) {
	canonical, err := canonicalize(payload)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(canonical)
	if err != nil {
		return "", fmt.Errorf("content checksum: marshal: %w", err)
	}
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func canonicalize(payload any) (any, error) {
	switch v := payload.(type) {
	case []CollectionFile:
		return canonicalizeFiles(v), nil
	case []CollectionPayload:
		files := make([]CollectionFile, len(v))
		for i, p := range v {
			files[i] = CollectionFile{Collection: p.Collection, Records: p.Records}
		}
		return canonicalizeFiles(files), nil
	default:
		return nil, fmt.Errorf("content checksum: unsupported payload type %T", payload)
	}
}

func canonicalizeFiles(files []CollectionFile) []CollectionFile {
	out := make([]CollectionFile, len(files))
	copy(out, files)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Collection < out[j].Collection
	})
	for i := range out {
		recs := out[i].Records
		sort.Slice(recs, func(a, b int) bool {
			ida, _ := recs[a]["id"].(string)
			idb, _ := recs[b]["id"].(string)
			return ida < idb
		})
		for _, rec := range recs {
			keys := make([]string, 0, len(rec))
			for k := range rec {
				keys = append(keys, k)
			}
			sort.Strings(keys)
		}
		out[i].Records = recs
	}
	return out
}
