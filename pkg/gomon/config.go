package gomon

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Job is one named entry in gomon.{json,toml,yaml}.
type Job struct {
	Exec     ExecValue `json:"exec"`
	Parallel bool      `json:"parallel"`
	Ignore   []string  `json:"ignore"`
	Watch    []string  `json:"watch"`
	Ext      []string  `json:"ext"`
}

// ExecValue is either one shell string or several (for concurrent/sequential runs).
type ExecValue struct {
	Parts []string
}

func (e *ExecValue) UnmarshalJSON(b []byte) error {
	b = trimSpaceBytes(b)
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		e.Parts = []string{s}
		return nil
	}
	if b[0] == '[' {
		var ss []string
		if err := json.Unmarshal(b, &ss); err != nil {
			return err
		}
		e.Parts = ss
		return nil
	}
	return fmt.Errorf("exec must be string or array of strings")
}

func trimSpaceBytes(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

// LoadConfig searches dir for gomon.json, gomon.yaml, gomon.yml, or gomon.toml
// (in that order) and returns all jobs plus the path of the file that was loaded.
func LoadConfig(dir string) (map[string]Job, string, error) {
	candidates := []string{
		filepath.Join(dir, "gomon.json"),
		filepath.Join(dir, "gomon.yaml"),
		filepath.Join(dir, "gomon.yml"),
		filepath.Join(dir, "gomon.toml"),
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, "", err
		}
		jobs, err := ParseConfigFile(path, data)
		if err != nil {
			return nil, "", fmt.Errorf("%s: %w", path, err)
		}
		return jobs, path, nil
	}
	return nil, "", fmt.Errorf("no gomon.json, gomon.yaml, or gomon.toml in %s", dir)
}

// ParseConfigFile parses data using the file extension of path (.json, .yaml, .yml, .toml).
func ParseConfigFile(path string, data []byte) (map[string]Job, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return parseJSONConfig(data)
	case ".yaml", ".yml":
		return parseSimpleYAMLConfig(data)
	case ".toml":
		return parseSimpleTOMLConfig(data)
	default:
		return nil, fmt.Errorf("unsupported config extension %q", ext)
	}
}

func parseJSONConfig(data []byte) (map[string]Job, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	out := make(map[string]Job, len(raw))
	for name, chunk := range raw {
		var j Job
		if err := json.Unmarshal(chunk, &j); err != nil {
			return nil, fmt.Errorf("job %q: %w", name, err)
		}
		out[name] = j
	}
	return out, nil
}

// ApplyDefaults fills in exec and watch when a job is empty or partially set.
func ApplyDefaults(j *Job) {
	if len(j.Exec.Parts) == 0 && len(j.Watch) == 0 {
		j.Exec.Parts = []string{"go run ."}
		j.Watch = []string{"./**"}
		return
	}
	if len(j.Watch) > 0 && len(j.Exec.Parts) == 0 {
		j.Exec.Parts = []string{"go run ."}
	}
	for i := range j.Watch {
		switch j.Watch[i] {
		case ".", "./":
			j.Watch[i] = "./**"
		}
	}
}
