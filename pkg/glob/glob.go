package glob

import (
	"path/filepath"
	"strings"
)

func MatchPattern(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)
	if !strings.Contains(pattern, "**") {
		ok, err := filepath.Match(pattern, path)
		return err == nil && ok
	}
	return matchDoubleStar(pattern, path)
}

func matchDoubleStar(pattern, path string) bool {
	parts := strings.Split(pattern, "**")
	if len(parts) < 2 {
		return false
	}
	prefix := normalizeGlobRoot(parts[0])
	suffix := strings.TrimPrefix(strings.Join(parts[1:], "**"), "/")
	suffix = normalizeGlobRoot(suffix)

	if prefix != "" {
		if !hasPrefixPath(path, prefix) {
			return false
		}
		path = trimPrefixPath(path, prefix)
	}

	if suffix == "" {
		return true
	}

	if strings.Contains(suffix, "**") {
		for path != "" && path != "." {
			if matchDoubleStar(suffix, path) {
				return true
			}
			if i := strings.Index(path, "/"); i >= 0 {
				path = path[i+1:]
			} else {
				break
			}
		}
		return false
	}

	return strings.HasSuffix(path, suffix) || path == suffix ||
		strings.HasSuffix(path, "/"+suffix)
}

func normalizeGlobRoot(p string) string {
	p = strings.TrimSpace(filepath.ToSlash(p))
	p = strings.TrimSuffix(p, "/")
	if p == "." || p == "" {
		return ""
	}
	return p
}

func hasPrefixPath(path, prefix string) bool {
	if prefix == "" {
		return true
	}
	if path == prefix {
		return true
	}
	return strings.HasPrefix(path, prefix+"/")
}

func trimPrefixPath(path, prefix string) string {
	if prefix == "" {
		return path
	}
	if path == prefix {
		return "."
	}
	if strings.HasPrefix(path, prefix+"/") {
		return path[len(prefix)+1:]
	}
	return path
}
