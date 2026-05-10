package htmltpl

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"
)

// DefaultFuncMap returns safe, boring helpers suitable for most HTML pages.
// Keys are stable; add app-specific funcs via [MergeFuncs].
func DefaultFuncMap() template.FuncMap {
	return template.FuncMap{
		"join":             strings.Join,
		"lower":            strings.ToLower,
		"upper":            strings.ToUpper,
		"trim":             strings.TrimSpace,
		"hasPrefix":        strings.HasPrefix,
		"hasSuffix":        strings.HasSuffix,
		"contains":         strings.Contains,
		"formatTime":       formatTimeRFC3339,
		"formatTimeLayout": formatTimeLayout,
		"formatMoneyCents": formatMoneyCents,
		"add":              addInt,
		"dict":             dictPairs,
	}
}

// MergeFuncs copies base then overlays extra (extra wins on key collision).
func MergeFuncs(base, extra template.FuncMap) template.FuncMap {
	out := template.FuncMap{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func formatTimeRFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// formatTimeLayout formats t with layout (empty layout defaults to RFC3339).
func formatTimeLayout(t time.Time, layout string) string {
	if t.IsZero() {
		return ""
	}
	if strings.TrimSpace(layout) == "" {
		layout = time.RFC3339
	}
	return t.UTC().Format(layout)
}

// formatMoneyCents renders cents as a fixed USD-style decimal (e.g. 1099 → "10.99").
// Negative values are supported. For other currencies, add your own template func.
func formatMoneyCents(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	whole := cents / 100
	frac := cents % 100
	s := strconv.FormatInt(whole, 10) + "." + fmt.Sprintf("%02d", frac)
	if neg {
		return "-" + s
	}
	return s
}

func addInt(a, b int) int { return a + b }

// dictPairs builds map[string]any from alternating key string, value pairs.
func dictPairs(v ...any) (map[string]any, error) {
	if len(v)%2 != 0 {
		return nil, fmt.Errorf("dict: want even number of arguments, got %d", len(v))
	}
	m := make(map[string]any, len(v)/2)
	for i := 0; i < len(v); i += 2 {
		k, ok := v[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict: key at %d must be string", i)
		}
		m[k] = v[i+1]
	}
	return m, nil
}
