package gomon

import (
	"reflect"
	"testing"
)

func TestParseJSONConfig(t *testing.T) {
	data := []byte(`{
  "dev": { "exec": "go run .", "watch": ["."] },
  "all": { "exec": ["go vet ./...", "go build"], "parallel": true, "watch": ["./**"] }
}`)
	jobs, err := ParseConfigFile("fixture.json", data)
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs["dev"].Exec.Parts) != 1 || jobs["dev"].Exec.Parts[0] != "go run ." {
		t.Fatalf("dev exec: %#v", jobs["dev"].Exec.Parts)
	}
	if !jobs["all"].Parallel || len(jobs["all"].Exec.Parts) != 2 {
		t.Fatalf("all: %#v", jobs["all"])
	}
}

func TestParseSimpleYAML(t *testing.T) {
	data := []byte(`
dev:
  exec: go run .
  watch:
    - ./
build:
  exec: go build .
`)
	jobs, err := ParseConfigFile("fixture.yaml", data)
	if err != nil {
		t.Fatal(err)
	}
	if jobs["dev"].Exec.Parts[0] != "go run ." {
		t.Fatalf("dev: %#v", jobs["dev"])
	}
	if jobs["build"].Exec.Parts[0] != "go build ." {
		t.Fatalf("build: %#v", jobs["build"])
	}
}

func TestParseSimpleTOML(t *testing.T) {
	data := `
[dev]
exec = "go run ."
watch = ["."]

[build]
exec = "go build ."
`
	jobs, err := ParseConfigFile("fixture.toml", []byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(jobs["dev"].Watch, []string{"."}) {
		t.Fatalf("watch: %#v", jobs["dev"].Watch)
	}
}

func TestMatchPattern(t *testing.T) {
	cases := []struct {
		pat  string
		path string
		want bool
	}{
		{"./**", "foo.go", true},
		{"./**", "cmd/foo.go", true},
		{"pkg/**", "pkg/a.go", true},
		{"pkg/**", "other/a.go", false},
		{"*.go", "main.go", true},
		{"*.go", "cmd/main.go", false},
		{".git/**", ".git/HEAD", true},
	}
	for _, tc := range cases {
		if got := MatchPattern(tc.pat, tc.path); got != tc.want {
			t.Errorf("MatchPattern(%q, %q) = %v, want %v", tc.pat, tc.path, got, tc.want)
		}
	}
}
