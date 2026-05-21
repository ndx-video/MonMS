package validate

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	gohtml "golang.org/x/net/html"
)

// templateDirectiveRE strips Go template directives before HTML tokenization.
// (?s) dotall flag allows . to match newlines — required for multi-line {{range}} blocks.
var templateDirectiveRE = regexp.MustCompile(`(?s)\{\{.*?\}\}`)

var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true,
	"embed": true, "hr": true, "img": true, "input": true,
	"link": true, "meta": true, "param": true, "source": true,
	"track": true, "wbr": true,
}

// ValidateTemplate performs a dry-run parse mirroring the production ssr.go loader (D-37).
// It uses template.ParseFiles(layoutPath, filePath) — identical two-arg call, layout first.
func ValidateTemplate(wsAbs, filePath string) error {
	layoutPath := filepath.Join(wsAbs, "templates", "layouts", "base.gohtml")
	if _, err := template.ParseFiles(layoutPath, filePath); err != nil {
		return fmt.Errorf("%s: template parse error: %w", filepath.Base(filePath), err)
	}
	return nil
}

// ValidateHTML checks well-formedness of HTML content using a tag-balance stack.
// Template directives are stripped before tokenization to avoid false positives.
func ValidateHTML(filePath string, content []byte) error {
	stripped := templateDirectiveRE.ReplaceAll(content, []byte(" "))
	z := gohtml.NewTokenizer(bytes.NewReader(stripped))

	var stack []string
	var errs []string

	for {
		tt := z.Next()
		switch tt {
		case gohtml.ErrorToken:
			if z.Err() == io.EOF {
				goto done
			}
			errs = append(errs, fmt.Sprintf("tokenizer error: %v", z.Err()))
			goto done
		case gohtml.StartTagToken:
			rawName, selfClose := z.TagName()
			name := string(rawName)
			if !selfClose && !voidElements[name] {
				stack = append(stack, name)
			}
		case gohtml.EndTagToken:
			rawName, _ := z.TagName()
			name := string(rawName)
			if len(stack) == 0 {
				errs = append(errs, fmt.Sprintf("unexpected </%s>: no open tag", name))
			} else if stack[len(stack)-1] != name {
				errs = append(errs, fmt.Sprintf("mismatched tag: open <%s>, close </%s>", stack[len(stack)-1], name))
			} else {
				stack = stack[:len(stack)-1]
			}
		}
	}
done:
	for _, tag := range stack {
		errs = append(errs, fmt.Sprintf("unclosed <%s>", tag))
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s: HTML structure:\n  %s", filepath.Base(filePath), strings.Join(errs, "\n  "))
	}
	return nil
}

// ValidateFiles orchestrates template and HTML validation for a list of files,
// accumulating all errors before returning (continue-on-error pattern from sync.go).
// T-02-01: file paths are verified to be under wsAbs before reading.
func ValidateFiles(wsAbs string, files []string) error {
	wsAbs = filepath.Clean(wsAbs)
	var errs []string

	for _, f := range files {
		// T-02-01: path traversal guard — reject files outside workspace.
		cleanF := filepath.Clean(f)
		rel, err := filepath.Rel(wsAbs, cleanF)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			errs = append(errs, fmt.Sprintf("%s: refusing to read file outside workspace", filepath.Base(f)))
			continue
		}

		content, err := os.ReadFile(cleanF)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: read error: %v", filepath.Base(f), err))
			continue
		}
		if err := ValidateTemplate(wsAbs, cleanF); err != nil {
			errs = append(errs, err.Error())
		}
		if err := ValidateHTML(cleanF, content); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation failed:\n%s", strings.Join(errs, "\n"))
	}
	return nil
}
