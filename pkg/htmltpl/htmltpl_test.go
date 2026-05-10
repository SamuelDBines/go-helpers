package htmltpl

import (
	"bytes"
	"embed"
	"html/template"
	"strings"
	"testing"
	"time"
)

//go:embed testdata/*.tmpl
var testFS embed.FS

func TestParseFSAndRender(t *testing.T) {
	s, err := ParseFS(testFS, nil,
		"testdata/layout.tmpl",
		"testdata/page.tmpl",
	)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	data := struct {
		Title string
		Now   time.Time
		Cents int64
	}{
		Title: "Hello",
		Now:   time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
		Cents: 1099,
	}
	if err := s.WriteTemplate(&buf, "layout", data); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Hello") {
		t.Fatalf("missing title: %q", out)
	}
	if !strings.Contains(out, "2026-05-01T12:00:00Z") && !strings.Contains(out, "2026-05-01T12:00:00") {
		t.Fatalf("missing formatted time: %q", out)
	}
	if !strings.Contains(out, "10.99") {
		t.Fatalf("missing money: %q", out)
	}
}

func TestMergeFuncsOverride(t *testing.T) {
	custom := template.FuncMap{
		"upper": func(s string) string { return "X" + s },
	}
	s, err := ParseFS(testFS, custom, "testdata/page_only.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := s.WriteTemplate(&buf, "only", struct{ Name string }{Name: "ab"}); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); !strings.Contains(got, "Xab") {
		t.Fatalf("custom upper not applied: %q", got)
	}
}

func TestDict(t *testing.T) {
	s, err := ParseFS(testFS, nil, "testdata/dict.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := s.WriteTemplate(&buf, "d", nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "v1") {
		t.Fatalf("dict render: %q", buf.String())
	}
}
