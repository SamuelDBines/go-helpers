package yamlw

import (
	"fmt"
	"sort"
	"strings"
)

type Builder struct {
	b      strings.Builder
	indent int
}

func New() *Builder { return &Builder{} }

func (y *Builder) String() string { return y.b.String() }

func (y *Builder) Indent(fn func()) {
	y.indent++
	fn()
	y.indent--
}

func (y *Builder) line(s string) {
	for i := 0; i < y.indent; i++ {
		y.b.WriteString("  ") // 2 spaces
	}
	y.b.WriteString(s)
	y.b.WriteByte('\n')
}

func (y *Builder) KV(key string, val any) {
	// key: <scalar>
	if isNil(val) {
		y.line(fmt.Sprintf("%s: null", key))
		return
	}
	if isScalar(val) {
		y.line(fmt.Sprintf("%s: %s", key, scalar(val)))
		return
	}

	// key:
	y.line(fmt.Sprintf("%s:", key))
	y.Indent(func() { y.Any(val) })
}

func (y *Builder) Map(key string, fn func()) {
	y.line(key + ":")
	y.Indent(fn)
}

func (y *Builder) List(key string, items []any) {
	y.line(key + ":")
	y.Indent(func() {
		for _, it := range items {
			y.Item(it)
		}
	})
}

func (y *Builder) Item(val any) {
	// - <scalar> OR
	// - <nested>
	if isNil(val) {
		y.line("- null")
		return
	}
	if isScalar(val) {
		y.line("- " + scalar(val))
		return
	}
	y.line("-")
	y.Indent(func() { y.Any(val) })
}

func (y *Builder) Any(v any) {
	switch t := v.(type) {
	case map[string]string:
		writeStringMap(y, t)
	case map[string]any:
		writeAnyMap(y, t)
	case []string:
		for _, s := range t {
			y.line("- " + scalar(s))
		}
	case []any:
		for _, it := range t {
			y.Item(it)
		}
	case Marshaler:
		t.YAML(y)
	default:
		// fall back to scalar-ish fmt
		y.line(scalar(fmt.Sprint(v)))
	}
}

// Marshaler lets your structs write themselves with the builder.
type Marshaler interface {
	YAML(y *Builder)
}

// ---------- helpers ----------

func isNil(v any) bool { return v == nil }

func isScalar(v any) bool {
	switch v.(type) {
	case string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	default:
		return false
	}
}

func (y *Builder) Line(s string) { y.line(s) }

func scalar(v any) string {
	switch t := v.(type) {
	case string:
		return quoteIfNeeded(t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		// numbers, etc
		return fmt.Sprint(v)
	}
}

func quoteIfNeeded(s string) string {
	// Safe-ish YAML scalar quoting for k8s fields
	if s == "" ||
		strings.ContainsAny(s, ":\n#{}[]&*!|>'\"%@`") ||
		strings.HasPrefix(s, " ") || strings.HasSuffix(s, " ") ||
		strings.HasPrefix(s, "-") || strings.HasPrefix(s, "?") ||
		strings.HasPrefix(s, "*") || strings.HasPrefix(s, "&") {
		// s = strings.ReplaceAll(s, `\`, `\\`)
		// s = strings.ReplaceAll(s, `"`, `\"`)
		return `"` + s + `"`
	}
	return s
}

func writeStringMap(y *Builder, m map[string]string) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys) // deterministic output
	for _, k := range keys {
		y.KV(k, m[k])
	}
}

func writeAnyMap(y *Builder, m map[string]any) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		y.KV(k, m[k])
	}
}
