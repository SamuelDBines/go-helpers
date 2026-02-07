package filestore

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type Store struct {
	Root string
}

type WriteOptions struct {
	Perm fs.FileMode // file perm, default 0644
}

func New(root string) Store {
	return Store{Root: root}
}

// Abs resolves a path under Root (unless already absolute).
func (s Store) Abs(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(s.Root, p)
}

// EnsureDirForFile creates parent dirs for a file path.
func (s Store) EnsureDirForFile(p string, perm fs.FileMode) error {
	abs := s.Abs(p)
	dir := filepath.Dir(abs)
	return os.MkdirAll(dir, perm)
}

// Exists checks whether a file/dir exists.
func (s Store) Exists(p string) bool {
	_, err := os.Stat(s.Abs(p))
	return err == nil
}

// Read reads full file contents.
func (s Store) Read(p string) ([]byte, error) {
	return os.ReadFile(s.Abs(p))
}

// ReadString reads file as string.
func (s Store) ReadString(p string) (string, error) {
	b, err := s.Read(p)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Write writes bytes to file, creating parent dirs.
func (s Store) Write(p string, data []byte, opts ...WriteOptions) error {
	perm := fs.FileMode(0644)
	if len(opts) > 0 && opts[0].Perm != 0 {
		perm = opts[0].Perm
	}
	if err := s.EnsureDirForFile(p, 0755); err != nil {
		return err
	}
	return os.WriteFile(s.Abs(p), data, perm)
}

// WriteString writes a string.
func (s Store) WriteString(p string, data string, opts ...WriteOptions) error {
	return s.Write(p, []byte(data), opts...)
}

// Append appends bytes to a file (creates file + parent dirs if needed).
func (s Store) Append(p string, data []byte, opts ...WriteOptions) error {
	perm := fs.FileMode(0644)
	if len(opts) > 0 && opts[0].Perm != 0 {
		perm = opts[0].Perm
	}

	if err := s.EnsureDirForFile(p, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(s.Abs(p), os.O_CREATE|os.O_WRONLY|os.O_APPEND, perm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

// AppendString appends a string.
func (s Store) AppendString(p string, data string, opts ...WriteOptions) error {
	return s.Append(p, []byte(data), opts...)
}

// Delete deletes a file. If missing, it’s a no-op.
func (s Store) Delete(p string) error {
	err := os.Remove(s.Abs(p))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// DeleteDir deletes a directory recursively. If missing, no-op.
func (s Store) DeleteDir(p string) error {
	err := os.RemoveAll(s.Abs(p))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// Copy copies file from src to dst (creates dst parent dirs).
func (s Store) Copy(src, dst string, perm fs.FileMode) error {
	if perm == 0 {
		perm = 0644
	}
	in, err := os.Open(s.Abs(src))
	if err != nil {
		return err
	}
	defer in.Close()

	if err := s.EnsureDirForFile(dst, 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(s.Abs(dst), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// ListDir returns entry names in a directory.
func (s Store) ListDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(s.Abs(dir))
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out, nil
}

// Walk walks files under a directory.
func (s Store) Walk(dir string, fn func(rel string, d fs.DirEntry) error) error {
	root := s.Abs(dir)
	return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(s.Root, p)
		if rerr != nil {
			return rerr
		}
		return fn(rel, d)
	})
}

// --- JSON helpers ---

func (s Store) WriteJSON(p string, v any, pretty bool) error {
	var (
		b   []byte
		err error
	)
	if pretty {
		b, err = json.MarshalIndent(v, "", "  ")
	} else {
		b, err = json.Marshal(v)
	}
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return s.Write(p, b, WriteOptions{Perm: 0644})
}

func (s Store) ReadJSON(p string, out any) error {
	b, err := s.Read(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

// --- “format” helpers (no parsing libs) ---

func (s Store) WriteYAML(p string, yamlText string) error {
	if len(yamlText) == 0 || yamlText[len(yamlText)-1] != '\n' {
		yamlText += "\n"
	}
	return s.WriteString(p, yamlText, WriteOptions{Perm: 0644})
}

func (s Store) ReadYAML(p string) (string, error) {
	return s.ReadString(p)
}

func (s Store) WriteHTML(p string, html string) error {
	return s.WriteString(p, html, WriteOptions{Perm: 0644})
}
