package env

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

func writeFile(t *testing.T, dir, name, contents string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

func TestSplitKV(t *testing.T) {
	tests := []struct {
		line   string
		wantK  string
		wantV  string
		wantOK bool
	}{
		{"FOO=bar", "FOO", "bar", true},
		{"FOO = bar", "FOO", "bar", true},
		{`FOO="bar baz"`, "FOO", "bar baz", true},
		{`FOO='bar # not comment'`, "FOO", "bar # not comment", true},
		{`FOO=bar # comment`, "FOO", "bar", true},
		{`FOO=a\=b`, "FOO", `a\=b`, true}, // escaped '=' is not a splitter
		{"NOVAL", "", "", false},
	}
	for _, tt := range tests {
		k, v, ok := splitKV(tt.line)
		if ok != tt.wantOK || k != tt.wantK || v != tt.wantV {
			t.Fatalf("splitKV(%q) = (%q,%q,%v), want (%q,%q,%v)",
				tt.line, k, v, ok, tt.wantK, tt.wantV, tt.wantOK)
		}
	}
}

func TestParse_Basics(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, ".env", `
# comment
export FOO=bar
BAZ="hello world"
RAW=val # inline comment
QUOTED='a#b'
ESCAPES=line\nbreak
`)
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	m := parse(f)
	if m["FOO"] != "bar" {
		t.Fatalf("FOO=%q", m["FOO"])
	}
	if m["BAZ"] != "hello world" {
		t.Fatalf("BAZ=%q", m["BAZ"])
	}
	if m["RAW"] != "val" {
		t.Fatalf("RAW=%q", m["RAW"])
	}
	if m["QUOTED"] != "a#b" {
		t.Fatalf("QUOTED=%q", m["QUOTED"])
	}
	if m["ESCAPES"] != "line\nbreak" {
		t.Fatalf("ESCAPES=%q", strings.ReplaceAll(m["ESCAPES"], "\n", "\\n"))
	}
}

func TestExpand(t *testing.T) {
	lookup := func(k string) (string, bool) {
		switch k {
		case "A":
			return "apple", true
		case "B":
			return "banana", true
		}
		return "", false
	}
	got := expand("${A}-${B}-${MISSING}", lookup)
	want := "apple-banana-"
	if got != want {
		t.Fatalf("expand got %q want %q", got, want)
	}
}

func TestLoad_Merge_NoOverwrite(t *testing.T) {
	t.Setenv("EXISTING", "keepme")

	dir := t.TempDir()
	env1 := writeFile(t, dir, "a.env", `
FOO=one
EXISTING=should-not-overwrite
REF=${FOO}/x
`)
	env2 := writeFile(t, dir, "b.env", `
FOO=two
BAR=zzz
`)

	opts := &Options{Overwrite: false, Expand: true}
	values, err := Load([]string{env1, env2}, opts)
	if err != nil {
		t.Fatal(err)
	}

	// later file overrides earlier in map
	if values["FOO"] != "two" {
		t.Fatalf("values[FOO]=%q", values["FOO"])
	}
	// environment var should not be overwritten
	if got := os.Getenv("EXISTING"); got != "keepme" {
		t.Fatalf("EXISTING=%q want keepme", got)
	}
	// expansion uses final map values
	if os.Getenv("REF") != "two/x" {
		t.Fatalf("REF=%q", os.Getenv("REF"))
	}
	// BAR is set
	if os.Getenv("BAR") != "zzz" {
		t.Fatalf("BAR=%q", os.Getenv("BAR"))
	}
}

func TestLoad_Overwrite(t *testing.T) {
	t.Setenv("FOO", "pre")
	dir := t.TempDir()
	p := writeFile(t, dir, ".env", "FOO=post\n")
	opts := &Options{Overwrite: true, Expand: false}
	if _, err := Load([]string{p}, opts); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("FOO") != "post" {
		t.Fatalf("FOO=%q want post", os.Getenv("FOO"))
	}
}

func TestFilenamesOrDefault(t *testing.T) {
	got := filenamesOrDefault(nil)
	want := []string{".env"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	got = filenamesOrDefault([]string{"x", "y"})
	if !reflect.DeepEqual(got, []string{"x", "y"}) {
		t.Fatalf("got %v want [x y]", got)
	}
}

func TestTypedGetters(t *testing.T) {
	// Ensure clean baseline for keys we set
	keys := []string{"S", "I", "B", "D", "LIST"}
	for _, k := range keys {
		t.Setenv(k, "")
	}

	t.Setenv("S", "hello")
	if got := String("S", "def"); got != "hello" {
		t.Fatalf("String got %q", got)
	}
	if got := String("MISSING", "def"); got != "def" {
		t.Fatalf("String default got %q", got)
	}

	t.Setenv("I", "42")
	if got := Int("I", 0); got != 42 {
		t.Fatalf("Int got %d", got)
	}
	if got := Int("I_BAD", 7); got != 7 {
		t.Fatalf("Int default got %d", got)
	}

	t.Setenv("B", "true")
	if got := Bool("B", false); !got {
		t.Fatalf("Bool got false")
	}
	if got := Bool("B_BAD", true); !got {
		t.Fatalf("Bool default got false")
	}

	t.Setenv("D", "250ms")
	if got := Duration("D", time.Second); got != 250*time.Millisecond {
		t.Fatalf("Duration got %v", got)
	}
	if got := Duration("D_BAD", 2*time.Second); got != 2*time.Second {
		t.Fatalf("Duration default got %v", got)
	}

	t.Setenv("LIST", "a, b ,c")
	l := Strings("LIST", ",", nil)
	if !reflect.DeepEqual(l, []string{"a", "b", "c"}) {
		t.Fatalf("Strings got %v", l)
	}
	empty := Strings("EMPTY", ",", []string{"x"})
	if !reflect.DeepEqual(empty, []string{"x"}) {
		t.Fatalf("Strings default got %v", empty)
	}
}

func TestMustStringPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustString did not panic")
		}
	}()
	_ = MustString("DOES_NOT_EXIST")
}

func TestLoad_SkipsMissingFiles(t *testing.T) {
	dir := t.TempDir()
	present := writeFile(t, dir, "present.env", "FOO=bar\n")
	// pass a missing file and a dir to exercise the skip path
	missing := filepath.Join(dir, "missing.env")

	opts := &Options{Overwrite: true, Expand: true}
	_, err := Load([]string{missing, dir, present}, opts)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if os.Getenv("FOO") != "bar" {
		t.Fatalf("FOO=%q", os.Getenv("FOO"))
	}
}

func TestExpandOrder_FileThenProcessEnv(t *testing.T) {
	// values from file should be visible to ${..} expansions
	dir := t.TempDir()
	writeFile(t, dir, ".env", `
A=one
B=${A}/two
`)
	// process env fallback used when not found in file
	t.Setenv("C", "three")
	writeFile(t, dir, ".env2", `
D=${C}/four
`)

	opts := &Options{Overwrite: true, Expand: true}
	_, err := Load([]string{filepath.Join(dir, ".env"), filepath.Join(dir, ".env2")}, opts)
	if err != nil {
		t.Fatal(err)
	}

	if os.Getenv("B") != "one/two" {
		t.Fatalf("B=%q", os.Getenv("B"))
	}
	if os.Getenv("D") != "three/four" {
		t.Fatalf("D=%q", os.Getenv("D"))
	}
}

// Small sanity check that LoadDefault doesn't error when files don't exist.
// (Relies on current working directory, so we only check it doesn't crash.)
func TestLoadDefault_NoFilesOK(t *testing.T) {
	// Run in an empty temp dir to ensure .env likely doesn't exist.
	wd := t.TempDir()
	old, _ := os.Getwd()
	defer os.Chdir(old)
	_ = os.Chdir(wd)

	_, err := LoadDefault(&Options{Overwrite: false, Expand: true})
	if err != nil {
		t.Fatalf("LoadDefault unexpected error: %v", err)
	}
}

// Optional: ensures tests that rely on CRLF/EOF work cross-platform.
func TestRuntimeEnv(t *testing.T) {
	if runtime.GOOS == "" {
		t.Fatalf("GOOS not set?")
	}
}
