package validate_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/monms/monms/internal/testutil"
	"github.com/monms/monms/internal/validate"
)

func TestValidateTemplate(t *testing.T) {
	ws := testutil.NewSite(t)
	pagePath := filepath.Join(ws, "templates/press/index.gohtml")
	testutil.WriteFile(t, pagePath, `{{define "body"}}<h1>Press</h1>{{end}}`)

	if err := validate.ValidateTemplate(ws, pagePath); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidateTemplateBadSyntax(t *testing.T) {
	ws := testutil.NewSite(t)
	pagePath := filepath.Join(ws, "templates/broken.gohtml")
	testutil.WriteFile(t, pagePath, `{{define "body"}}{{if}}{{end}}`)

	err := validate.ValidateTemplate(ws, pagePath)
	if err == nil {
		t.Fatal("expected template parse error, got nil")
	}
	if !strings.Contains(err.Error(), "template parse error") {
		t.Fatalf("expected 'template parse error' in message, got: %v", err)
	}
}

func TestValidateHTML(t *testing.T) {
	cases := []struct {
		name    string
		content string
		wantErr bool
	}{
		{"valid div", `<div><p>text</p></div>`, false},
		{"unclosed div", `<div><p>text</div>`, true},
		{"void elements ok", `<br><img src="x.png"><input type="text">`, false},
		{"template directives stripped", `{{define "body"}}<div></div>{{end}}`, false},
		{"multiline range directive", "{{range .Items}}\n<li>{{.Title}}</li>\n{{end}}", false},
		{"gt in directive", `{{if .Count > 0}}<span>ok</span>{{end}}`, false},
		{"mismatched close", `<div><span></div>`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validate.ValidateHTML("test.gohtml", []byte(tc.content))
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tc.wantErr, err)
			}
		})
	}
}

func TestValidateFilesAggregatesErrors(t *testing.T) {
	ws := testutil.NewSite(t)

	broken1 := filepath.Join(ws, "templates/b1.gohtml")
	broken2 := filepath.Join(ws, "templates/b2.gohtml")
	testutil.WriteFile(t, broken1, `{{define "body"}}{{if}}{{end}}`)
	testutil.WriteFile(t, broken2, `{{define "body"}}{{if}}{{end}}`)

	err := validate.ValidateFiles(ws, []string{broken1, broken2})
	if err == nil {
		t.Fatal("expected aggregated error, got nil")
	}
	if !strings.Contains(err.Error(), "b1.gohtml") {
		t.Errorf("expected b1.gohtml in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "b2.gohtml") {
		t.Errorf("expected b2.gohtml in error, got: %v", err)
	}
}
