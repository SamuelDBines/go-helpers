package htmltpl

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
)

const htmlUTF8 = "text/html; charset=utf-8"

// Set is a parsed [html/template] tree ready for [Set.ExecuteTemplate].
type Set struct {
	root *template.Template
}

// ParseFS parses named files from fsys into one template set.
// funcs is merged with [DefaultFuncMap] (your funcs override defaults on conflict).
// patterns are paths relative to fsys (same rules as [template.Template.ParseFS]).
func ParseFS(fsys fs.FS, funcs template.FuncMap, patterns ...string) (*Set, error) {
	if fsys == nil {
		return nil, fmt.Errorf("htmltpl: nil fs")
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("htmltpl: need at least one template path")
	}
	rootName := filepath.Base(patterns[0])
	if rootName == "." || rootName == "/" {
		rootName = "root"
	}
	merged := MergeFuncs(DefaultFuncMap(), funcs)
	root := template.New(rootName).Funcs(merged)
	t, err := root.ParseFS(fsys, patterns...)
	if err != nil {
		return nil, err
	}
	return &Set{root: t}, nil
}

// MustParseFS is like [ParseFS] but panics on error (startup / init only).
func MustParseFS(fsys fs.FS, funcs template.FuncMap, patterns ...string) *Set {
	s, err := ParseFS(fsys, funcs, patterns...)
	if err != nil {
		panic(err)
	}
	return s
}

// Template returns the underlying template for advanced use (Clone, etc.).
func (s *Set) Template() *template.Template {
	if s == nil {
		return nil
	}
	return s.root
}

// ExecuteTemplate renders the named template into w and sets HTML Content-Type.
func (s *Set) ExecuteTemplate(w http.ResponseWriter, name string, data any) error {
	if s == nil || s.root == nil {
		return fmt.Errorf("htmltpl: nil Set")
	}
	w.Header().Set("Content-Type", htmlUTF8)
	return s.root.ExecuteTemplate(w, name, data)
}

// Execute renders the default (root) template into w.
func (s *Set) Execute(w http.ResponseWriter, data any) error {
	if s == nil || s.root == nil {
		return fmt.Errorf("htmltpl: nil Set")
	}
	w.Header().Set("Content-Type", htmlUTF8)
	return s.root.Execute(w, data)
}

// WriteTemplate renders to an arbitrary [io.Writer] (tests, buffers) without HTTP headers.
func (s *Set) WriteTemplate(w io.Writer, name string, data any) error {
	if s == nil || s.root == nil {
		return fmt.Errorf("htmltpl: nil Set")
	}
	return s.root.ExecuteTemplate(w, name, data)
}
