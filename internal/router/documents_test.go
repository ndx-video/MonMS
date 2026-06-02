package router

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/testutil"
)

func TestMarkdownDocumentSSR(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, "templates/doc.gohtml"), `{{define "body"}}
<article><h1>{{.Doc.Title}}</h1>{{.Doc.HTMLBody}}</article>
{{end}}`)

	schemaJSON := `{
  "name": "articles",
  "type": "base",
  "editorial": true,
  "monms": { "source": "markdown", "root": "documents/articles" },
  "fields": [
    { "name": "id", "type": "text", "primaryKey": true, "required": true, "pattern": "^[a-z][a-z0-9_-]*$", "min": 1, "max": 120 },
    { "name": "title", "type": "text" },
    { "name": "slug", "type": "text" },
    { "name": "path", "type": "text" },
    { "name": "body", "type": "text" }
  ]
}`
	testutil.WriteFile(t, filepath.Join(ws, "schema/articles.json"), schemaJSON)
	testutil.WriteFile(t, filepath.Join(ws, "documents/articles/hello.md"), `---
title: Hello Doc
---
# Hello

Markdown **body**.
`)

	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	resp, err := http.Get(ts.URL + "/hello")
	if err != nil {
		t.Fatalf("GET /hello: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "Hello Doc") {
		t.Fatalf("body missing title: %.200s", text)
	}
	if !strings.Contains(text, "<strong>body</strong>") {
		t.Fatalf("body missing rendered markdown: %.200s", text)
	}
}

func TestGohtmlWinsOverMarkdownDocument(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, "templates/hello.gohtml"), `{{define "body"}}<h1>Template wins</h1>{{end}}`)
	testutil.WriteFile(t, filepath.Join(ws, "templates/doc.gohtml"), `{{define "body"}}<h1>{{.Doc.Title}}</h1>{{end}}`)

	schemaJSON := `{
  "name": "articles",
  "type": "base",
  "editorial": true,
  "monms": { "source": "markdown", "root": "documents/articles" },
  "fields": [
    { "name": "id", "type": "text", "primaryKey": true, "required": true, "pattern": "^[a-z][a-z0-9_-]*$", "min": 1, "max": 120 },
    { "name": "title", "type": "text" },
    { "name": "slug", "type": "text" },
    { "name": "body", "type": "text" }
  ]
}`
	testutil.WriteFile(t, filepath.Join(ws, "schema/articles.json"), schemaJSON)
	testutil.WriteFile(t, filepath.Join(ws, "documents/articles/hello.md"), `---
title: Markdown Title
---
Body
`)

	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	resp, err := http.Get(ts.URL + "/hello")
	if err != nil {
		t.Fatalf("GET /hello: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Template wins") {
		t.Fatalf("expected template precedence, got: %.200s", string(body))
	}
}

func TestDocSectionTemplateFunc(t *testing.T) {
	ws := testutil.NewSite(t)
	testutil.WriteFile(t, filepath.Join(ws, "templates/index.gohtml"), `{{define "body"}}
<h2>{{ docHeading "hello" 2 0 }}</h2>
<div>{{ docSection "hello" 2 0 }}</div>
{{end}}`)

	schemaJSON := `{
  "name": "articles",
  "type": "base",
  "editorial": true,
  "monms": { "source": "markdown", "root": "documents/articles" },
  "fields": [
    { "name": "id", "type": "text", "primaryKey": true, "required": true, "pattern": "^[a-z][a-z0-9_-]*$", "min": 1, "max": 120 },
    { "name": "title", "type": "text" },
    { "name": "slug", "type": "text" },
    { "name": "path", "type": "text" },
    { "name": "body", "type": "text" }
  ]
}`
	testutil.WriteFile(t, filepath.Join(ws, "schema/articles.json"), schemaJSON)
	testutil.WriteFile(t, filepath.Join(ws, "documents/articles/hello.md"), `---
title: Hello Doc
---
# Title

## Intro

Section **one** content.

## Outro

Second section.
`)

	ts, _, cleanup := startTestServer(t, ws, testServerOpts{isDev: true})
	defer cleanup()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	text := string(body)
	if !strings.Contains(text, "Intro") {
		t.Fatalf("missing heading: %.300s", text)
	}
	if !strings.Contains(text, "<strong>one</strong>") {
		t.Fatalf("missing section HTML: %.300s", text)
	}
	if strings.Contains(text, "Second section") {
		t.Fatalf("should not include second section: %.300s", text)
	}
}
