package documents

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// inferDoctreeTitle picks title from first ATX H1, else humanized basename.
func inferDoctreeTitle(doc ParsedDocument, relPath string) string {
	for _, line := range strings.Split(doc.Body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "#"))
		}
	}
	return humanizeBasename(relPath)
}

func humanizeBasename(relPath string) string {
	base := filepath.Base(relPath)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return strings.ReplaceAll(base, "-", " ")
}

// mergeDoctreeFrontmatter applies id, semantic fields, and timestamps for dt_* bindings.
func mergeDoctreeFrontmatter(meta map[string]any, doc ParsedDocument, collection, pathKey, filePath string, force bool) (map[string]any, bool) {
	title := ""
	if t, ok := meta["title"].(string); ok && strings.TrimSpace(t) != "" {
		title = strings.TrimSpace(t)
	} else {
		title = inferDoctreeTitle(doc, pathKey+".md")
	}

	newMeta, changed := mergeFrontmatterMeta(meta, collection, pathKey, title, force)

	tsCreate, tsMod := fileTimestamps(filePath)
	if v, ok := newMeta["ts_create"].(string); !ok || strings.TrimSpace(v) == "" {
		newMeta["ts_create"] = tsCreate
		changed = true
	}
	if v, ok := newMeta["ts_mod"].(string); !ok || strings.TrimSpace(v) == "" {
		newMeta["ts_mod"] = tsMod
		changed = true
	}

	return newMeta, changed
}

func fileTimestamps(path string) (create, mod string) {
	now := time.Now().UTC().Format(time.RFC3339)
	fi, err := os.Stat(path)
	if err != nil {
		return now, now
	}
	mod = fi.ModTime().UTC().Format(time.RFC3339)
	return mod, mod
}
