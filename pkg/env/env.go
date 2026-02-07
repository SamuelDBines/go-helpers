package env

import (
	"bufio"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	Overwrite bool
	Expand    bool
}

func filenamesOrDefault(filenames []string) []string {
	if len(filenames) == 0 {
		return []string{".env"}
	}
	return filenames
}

func parse(r *os.File) map[string]string {
	out := make(map[string]string)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export"))
		}
		k, v, ok := splitKV(line)
		if !ok || k == "" {
			continue
		}
		out[k] = v
	}
	return out
}

func splitKV(line string) (key, val string, ok bool) {
	// Find first unescaped '='
	i := -1
	esc := false
	for idx, r := range line {
		if r == '\\' {
			esc = !esc
			continue
		}
		if r == '=' && !esc {
			i = idx
			break
		}
		esc = false
	}
	if i == -1 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:i])
	val = strings.TrimSpace(line[i+1:])
	val = stripQuotes(val)
	val = trimInlineComment(val)
	val = unescape(val)
	return key, val, true
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func trimInlineComment(s string) string {
	// If value is quoted, comments were already handled in stripQuotes.
	// For unquoted values, treat ' #' as start of comment.
	if strings.Contains(s, "#") {
		// stop at first ' #' (hash preceded by space or start)
		for i := 0; i < len(s); i++ {
			if s[i] == '#' && (i == 0 || s[i-1] == ' ') {
				return strings.TrimSpace(s[:i])
			}
		}
	}
	return strings.TrimSpace(s)
}

func unescape(s string) string {
	// handle simple escapes for \n, \t, \\, \"
	replacer := strings.NewReplacer(
		`\\`, `\\`,
		`\n`, "\n",
		`\r`, "\r",
		`\t`, "\t",
		`\"`, `"`,
		`\'`, `'`,
	)
	return replacer.Replace(s)
}

func expand(s string, lookup func(string) (string, bool)) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '$' && i+1 < len(s) && s[i+1] == '{' {
			// ${VAR}
			j := i + 2
			for j < len(s) && s[j] != '}' {
				j++
			}
			if j < len(s) {
				name := s[i+2 : j]
				if v, ok := lookup(name); ok {
					b.WriteString(v)
				}
				i = j + 1
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func GetEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func MustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing required env %s", k)
	}
	return v
}

func String(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

func MustString(key string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	panic(errors.New("missing required env: " + key))
}

func Int(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return n
		}
	}
	return def
}

func Bool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
			return b
		}
	}
	return def
}

func Duration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(strings.TrimSpace(v)); err == nil {
			return d
		}
	}
	return def
}

func Strings(key string, sep string, def []string) []string {
	if v, ok := os.LookupEnv(key); ok {
		if strings.TrimSpace(v) == "" {
			return []string{}
		}
		parts := strings.Split(v, sep)
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	return def
}

func parseFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parse(f), nil
}

func LoadDefault(opt *Options) (map[string]string, error) {
	return Load([]string{".env", ".env.local"}, opt)
}

func Load(files []string, opt *Options) (map[string]string, error) {
	if opt == nil {
		opt = &Options{}
	}
	values := make(map[string]string)
	for _, f := range filenamesOrDefault(files) {
		if f == "" {
			continue
		}
		path, _ := filepath.Abs(f)
		fi, err := os.Stat(path)
		if err != nil || fi.IsDir() {
			print("skipping missing or directory env file: ", path, "\n")
			continue
		}
		m, err := parseFile(path)
		if err != nil {
			return nil, err
		}
		for k, v := range m {
			if _, exists := values[k]; !exists || opt.Overwrite {
				if opt.Expand {
					v = os.ExpandEnv(v)
				}
				values[k] = v
			}
		}
	}
	for k, v := range values {
		val := v
		if opt.Expand {
			val = expand(val, func(name string) (string, bool) {
				// expansion order: already-applied values -> process env
				if vv, ok := values[name]; ok {
					return vv, true
				}
				if vv, ok := os.LookupEnv(name); ok {
					return vv, true
				}
				return "", false
			})
		}
		if _, exists := os.LookupEnv(k); exists && !opt.Overwrite {
			continue
		}
		_ = os.Setenv(k, val)
	}
	return values, nil
}
