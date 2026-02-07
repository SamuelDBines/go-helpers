package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Options struct {
	Name     string
	Level    slog.Leveler
	UseColor bool
	JSON     bool
	Out      io.Writer
	TimeFn   func() time.Time
}

func New(opts Options) *slog.Logger {
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.Level == nil {
		opts.Level = slog.LevelInfo
	}
	if opts.TimeFn == nil {
		opts.TimeFn = time.Now
	}

	h := &Handler{
		out:      opts.Out,
		level:    opts.Level,
		name:     opts.Name,
		useColor: opts.UseColor,
		json:     opts.JSON,
		timeFn:   opts.TimeFn,
	}
	return slog.New(h)
}

type Handler struct {
	out      io.Writer
	level    slog.Leveler
	name     string
	useColor bool
	json     bool
	timeFn   func() time.Time

	attrs  []slog.Attr
	groups []string
}

func (h *Handler) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= h.level.Level()
}

func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	src := ""
	if r.Level <= slog.LevelDebug || r.Level >= slog.LevelWarn {
		src = formatSource(r.PC)
	}

	if h.json {
		return h.writeJSON(r, src)
	}
	return h.writeText(r, src)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	n := *h
	n.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &n
}

func (h *Handler) WithGroup(name string) slog.Handler {
	n := *h
	n.groups = append(append([]string{}, h.groups...), name)
	return &n
}

func (h *Handler) writeText(r slog.Record, src string) error {
	ts := h.timeFn().Format(time.StampMilli)

	level := levelLabel(r.Level, h.useColor)
	name := ""
	if h.name != "" {
		name = faint(h.useColor, "["+h.name+"] ")
	}

	msg := r.Message

	var b strings.Builder

	fmt.Fprintf(&b, "%s %s %s", faint(h.useColor, ts), level, name)
	if src != "" {
		fmt.Fprintf(&b, "%s ", faint(h.useColor, src))
	}
	b.WriteString(msg)

	all := make([]slog.Attr, 0, len(h.attrs)+8)
	all = append(all, h.attrs...)

	r.Attrs(func(a slog.Attr) bool {
		all = append(all, a)
		return true
	})

	// group prefix e.g. group1.group2.key
	prefix := strings.Join(h.groups, ".")
	for _, a := range all {
		// a = a.Resolve()
		k := a.Key
		if prefix != "" {
			k = prefix + "." + k
		}
		// spacing and formatting
		fmt.Fprintf(&b, " %s=%s", faint(h.useColor, k), formatValue(a.Value))
	}

	b.WriteByte('\n')

	_, err := io.WriteString(h.out, b.String())
	return err
}

func (h *Handler) writeJSON(r slog.Record, src string) error {
	var b strings.Builder
	ts := h.timeFn().Format(time.RFC3339Nano)

	b.WriteString(`{"time":`)
	b.WriteString(jsonString(ts))
	b.WriteString(`,"level":`)
	b.WriteString(jsonString(r.Level.String()))
	if h.name != "" {
		b.WriteString(`,"logger":`)
		b.WriteString(jsonString(h.name))
	}
	if src != "" {
		b.WriteString(`,"source":`)
		b.WriteString(jsonString(src))
	}
	b.WriteString(`,"msg":`)
	b.WriteString(jsonString(r.Message))

	// attrs
	b.WriteString(`,"attrs":{`)

	first := true
	prefix := strings.Join(h.groups, ".")

	emit := func(a slog.Attr) {
		// a = a.Resolve()
		k := a.Key
		if prefix != "" {
			k = prefix + "." + k
		}
		if !first {
			b.WriteByte(',')
		}
		first = false
		b.WriteString(jsonString(k))
		b.WriteByte(':')
		b.WriteString(jsonValue(a.Value))
	}

	for _, a := range h.attrs {
		emit(a)
	}
	r.Attrs(func(a slog.Attr) bool {
		emit(a)
		return true
	})

	b.WriteString("}}\n")
	_, err := io.WriteString(h.out, b.String())
	return err
}

// ---------- formatting helpers ----------

const (
	ansiReset        = "\x1b[0m"
	ansiFaint        = "\x1b[2m"
	ansiBrightRed    = "\x1b[91m"
	ansiBrightGreen  = "\x1b[92m"
	ansiBrightYellow = "\x1b[93m"
	ansiBrightBlue   = "\x1b[94m"
)

func faint(ok bool, s string) string {
	if !ok {
		return s
	}
	return ansiFaint + s + ansiReset
}

func color(ok bool, c, s string) string {
	if !ok {
		return s
	}
	return c + s + ansiReset
}

func levelLabel(lvl slog.Level, useColor bool) string {
	switch {
	case lvl <= slog.LevelDebug:
		return color(useColor, ansiBrightBlue, "DEBUG")
	case lvl < slog.LevelWarn:
		return color(useColor, ansiBrightGreen, "INFO ")
	case lvl < slog.LevelError:
		return color(useColor, ansiBrightYellow, "WARN ")
	default:
		return color(useColor, ansiBrightRed, "ERROR")
	}
}

func formatSource(pc uintptr) string {
	if pc == 0 {
		// If slog.Record.PC isn't set, we can attempt to find caller here,
		// but that adds overhead. Keep it simple: empty if not provided.
		return ""
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}

	file, line := fn.FileLine(pc)
	file = filepath.Base(file)

	funcName := fn.Name()
	if i := strings.LastIndex(funcName, "/"); i >= 0 {
		funcName = funcName[i+1:]
	}
	// keep just last segment after dot if you want shorter:
	// if j := strings.LastIndex(funcName, "."); j >= 0 { funcName = funcName[j+1:] }

	return fmt.Sprintf("%s:%d %s()", file, line, funcName)
}

func formatValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return fmt.Sprintf("%q", v.String())
	case slog.KindInt64:
		return fmt.Sprintf("%d", v.Int64())
	case slog.KindUint64:
		return fmt.Sprintf("%d", v.Uint64())
	case slog.KindFloat64:
		return fmt.Sprintf("%g", v.Float64())
	case slog.KindBool:
		return fmt.Sprintf("%t", v.Bool())
	case slog.KindTime:
		return v.Time().Format(time.RFC3339Nano)
	case slog.KindDuration:
		return v.Duration().String()
	default:
		// KindAny, KindGroup, etc.
		return v.String()
	}
}

// -------- minimal JSON string/value helpers --------

func jsonString(s string) string {
	// minimal escape: backslash + quote + newlines/tabs
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func jsonValue(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return jsonString(v.String())
	case slog.KindInt64:
		return fmt.Sprintf("%d", v.Int64())
	case slog.KindUint64:
		return fmt.Sprintf("%d", v.Uint64())
	case slog.KindFloat64:
		return fmt.Sprintf("%g", v.Float64())
	case slog.KindBool:
		if v.Bool() {
			return "true"
		}
		return "false"
	case slog.KindTime:
		return jsonString(v.Time().Format(time.RFC3339Nano))
	case slog.KindDuration:
		return jsonString(v.Duration().String())
	default:
		return jsonString(v.String())
	}
}
