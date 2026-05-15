package env

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func String(key string, def ...string) string {
	v, ok := os.LookupEnv(key)
	if ok && strings.TrimSpace(v) != "" {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	panic("missing env: " + key)
}

func Get(key string, def ...string) string {
	return String(key, def...)
}

func Int(key string, def ...int) int {
	if v, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic("invalid int env " + key + ": " + v)
		}
		return i
	}

	if len(def) > 0 {
		return def[0]
	}

	panic("missing env: " + key)
}

func IntDefault(key string, def int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

func Duration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func CommaList(key string) []string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return nil
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func Bool(key string, def ...bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		s := strings.TrimSpace(v)
		if s != "" {
			switch strings.ToLower(s) {
			case "1", "true", "yes", "on":
				return true
			default:
				return false
			}
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return false
}

func LoadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		i := strings.IndexByte(line, '=')
		if i <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		val := strings.TrimSpace(line[i+1:])
		val = strings.Trim(val, `"'`)
		if key == "" {
			continue
		}
		if _, set := os.LookupEnv(key); set {
			continue
		}
		if err := os.Setenv(key, val); err != nil {
			return fmt.Errorf("setenv %q: %w", key, err)
		}
	}
	return sc.Err()
}

func Load(path ...string) error {
	if len(path) == 0 {
		path = []string{".env"}
	}

	for _, p := range path {
		if err := LoadEnvFile(p); err != nil {
			return err
		}
	}
	return nil
}
