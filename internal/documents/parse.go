package documents

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v2"
)

// ParsedDocument is one markdown file with frontmatter metadata and body.
type ParsedDocument struct {
	FilePath string
	RelPath  string
	Meta     map[string]any
	Body     string
}

// ParseFile reads a markdown file and splits frontmatter from body.
func ParseFile(path string) (ParsedDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ParsedDocument{}, err
	}
	return ParseBytes(path, data)
}

// ParseBytes parses markdown bytes with optional YAML frontmatter.
func ParseBytes(filePath string, data []byte) (ParsedDocument, error) {
	meta := map[string]any{}
	body, err := frontmatter.Parse(bytes.NewReader(data), &meta)
	if err != nil {
		return ParsedDocument{}, fmt.Errorf("documents parse %s: %w", filePath, err)
	}
	return ParsedDocument{
		FilePath: filePath,
		Meta:     meta,
		Body:     strings.TrimSpace(string(body)),
	}, nil
}

// WriteFile writes a markdown file with YAML frontmatter and body.
func WriteFile(path string, meta map[string]any, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var buf bytes.Buffer
	if len(meta) > 0 {
		buf.WriteString("---\n")
		yamlBytes, err := yaml.Marshal(meta)
		if err != nil {
			return fmt.Errorf("documents encode frontmatter: %w", err)
		}
		buf.Write(yamlBytes)
		buf.WriteString("---\n")
	}
	if body != "" {
		if _, err := buf.WriteString(body); err != nil {
			return err
		}
		if !strings.HasSuffix(body, "\n") {
			if _, err := buf.WriteString("\n"); err != nil {
				return err
			}
		}
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
