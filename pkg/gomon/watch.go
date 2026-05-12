package gomon

import (
	"os"
	"path/filepath"
	"strings"
)

// SnapshotMTimes walks root and returns modification times (nanoseconds) for files
// matched by job watch/ignore/ext rules. Keys are slash-separated paths relative to root.
func SnapshotMTimes(root string, job Job) (map[string]int64, error) {
	out := make(map[string]int64)
	watch := job.Watch
	if len(watch) == 0 {
		return out, nil
	}
	defaultIgnores := []string{".git/**", "vendor/**", "node_modules/**"}
	ignores := append(append([]string{}, defaultIgnores...), job.Ignore...)

	err := filepath.WalkDir(root, func(full string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			base := filepath.Base(full)
			if base == ".git" || base == "vendor" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, full)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		for _, ign := range ignores {
			if MatchPattern(ign, rel) {
				return nil
			}
		}
		matched := false
		for _, w := range watch {
			if MatchPattern(w, rel) || MatchPattern(w, "./"+rel) {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}
		if len(job.Ext) > 0 {
			ext := strings.ToLower(filepath.Ext(full))
			ok := false
			for _, want := range job.Ext {
				want = strings.ToLower(want)
				if !strings.HasPrefix(want, ".") {
					want = "." + want
				}
				if ext == want {
					ok = true
					break
				}
			}
			if !ok {
				return nil
			}
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		out[rel] = info.ModTime().UnixNano()
		return nil
	})
	return out, err
}

func mapsEqualInt64(a, b map[string]int64) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
