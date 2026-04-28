package zipkit

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteZip(t *testing.T) {
	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.txt")
	p2 := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(p1, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p2, []byte("world"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err := WriteZip(&buf, []File{{Name: "a.txt", Path: p1}, {Name: "nested/b.txt", Path: p2}})
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() < 50 {
		t.Fatalf("zip too small: %d", buf.Len())
	}
}
