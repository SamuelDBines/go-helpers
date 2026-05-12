package gomon

import (
	"fmt"
	"strings"
)

func parseSimpleTOMLConfig(data []byte) (map[string]Job, error) {
	lines := strings.Split(string(data), "\n")
	jobs := make(map[string]Job)
	var cur string
	var j Job
	var pendingKey string

	flush := func() {
		if cur != "" {
			jobs[cur] = j
		}
	}

	for _, line := range lines {
		line = strings.TrimSpace(stripCommentTOML(line))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			flush()
			cur = strings.TrimSpace(line[1 : len(line)-1])
			j = Job{}
			pendingKey = ""
			continue
		}
		if pendingKey != "" {
			if strings.TrimSpace(line) == "]" {
				pendingKey = ""
				continue
			}
			if strings.HasPrefix(line, "\"") {
				s, rest, err := parseTOMLString(line)
				if err != nil {
					return nil, fmt.Errorf("job %q: %w", cur, err)
				}
				if strings.TrimSpace(rest) != "" && !strings.HasPrefix(strings.TrimSpace(rest), ",") {
					return nil, fmt.Errorf("job %q: trailing garbage after string in array", cur)
				}
				switch pendingKey {
				case "exec":
					j.Exec.Parts = append(j.Exec.Parts, s)
				case "ignore":
					j.Ignore = append(j.Ignore, s)
				case "watch":
					j.Watch = append(j.Watch, s)
				case "ext":
					j.Ext = append(j.Ext, s)
				}
				continue
			}
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("job %q: expected key = value, got %q", cur, line)
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "parallel":
			switch strings.ToLower(val) {
			case "true":
				j.Parallel = true
			case "false":
				j.Parallel = false
			default:
				return nil, fmt.Errorf("job %q: parallel must be true or false", cur)
			}
		case "exec":
			if strings.HasPrefix(val, "[") {
				pendingKey = "exec"
				j.Exec.Parts = nil
				rest := strings.TrimPrefix(val, "[")
				rest = strings.TrimSpace(rest)
				if strings.HasSuffix(rest, "]") {
					rest = strings.TrimSuffix(rest, "]")
					rest = strings.TrimSpace(rest)
					if rest != "" {
						parts, err := splitTOMLCommaStrings(rest)
						if err != nil {
							return nil, fmt.Errorf("job %q: exec array: %w", cur, err)
						}
						j.Exec.Parts = parts
					}
					pendingKey = ""
				}
			} else {
				s, _, err := parseTOMLString(val)
				if err != nil {
					return nil, fmt.Errorf("job %q: exec: %w", cur, err)
				}
				j.Exec.Parts = []string{s}
			}
		case "ignore", "watch", "ext":
			field := key
			if strings.HasPrefix(val, "[") {
				pendingKey = field
				switch field {
				case "ignore":
					j.Ignore = nil
				case "watch":
					j.Watch = nil
				case "ext":
					j.Ext = nil
				}
				rest := strings.TrimPrefix(val, "[")
				rest = strings.TrimSpace(rest)
				if strings.HasSuffix(rest, "]") {
					rest = strings.TrimSuffix(rest, "]")
					rest = strings.TrimSpace(rest)
					parts, err := splitTOMLCommaStrings(rest)
					if err != nil {
						return nil, fmt.Errorf("job %q: %s: %w", cur, field, err)
					}
					switch field {
					case "ignore":
						j.Ignore = parts
					case "watch":
						j.Watch = parts
					case "ext":
						j.Ext = parts
					}
					pendingKey = ""
				}
			} else {
				s, _, err := parseTOMLString(val)
				if err != nil {
					return nil, fmt.Errorf("job %q: %s: %w", cur, field, err)
				}
				switch field {
				case "ignore":
					j.Ignore = []string{s}
				case "watch":
					j.Watch = []string{s}
				case "ext":
					j.Ext = []string{s}
				}
			}
		default:
			return nil, fmt.Errorf("job %q: unknown key %q", cur, key)
		}
	}
	flush()
	if len(jobs) == 0 {
		return nil, fmt.Errorf("no [job] sections found")
	}
	return jobs, nil
}

func stripCommentTOML(line string) string {
	inStr := false
	escape := false
	for i, r := range line {
		if escape {
			escape = false
			continue
		}
		if r == '\\' && inStr {
			escape = true
			continue
		}
		if r == '"' {
			inStr = !inStr
			continue
		}
		if !inStr && r == '#' {
			return line[:i]
		}
	}
	return line
}

func parseTOMLString(s string) (value string, rest string, err error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, `"`) {
		return "", "", fmt.Errorf("expected quoted string, got %q", s)
	}
	var b strings.Builder
	i := 1
	for i < len(s) {
		c := s[i]
		if c == '\\' && i+1 < len(s) {
			b.WriteByte(s[i+1])
			i += 2
			continue
		}
		if c == '"' {
			return b.String(), strings.TrimSpace(s[i+1:]), nil
		}
		b.WriteByte(c)
		i++
	}
	return "", "", fmt.Errorf("unterminated string")
}

func splitTOMLCommaStrings(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	var out []string
	for len(strings.TrimSpace(s)) > 0 {
		s = strings.TrimSpace(s)
		if !strings.HasPrefix(s, `"`) {
			return nil, fmt.Errorf("expected quoted element in %q", s)
		}
		part, rest, err := parseTOMLString(s)
		if err != nil {
			return nil, err
		}
		out = append(out, part)
		s = strings.TrimSpace(rest)
		if s == "" {
			break
		}
		if strings.HasPrefix(s, ",") {
			s = strings.TrimSpace(s[1:])
			continue
		}
		return nil, fmt.Errorf("expected comma in %q", s)
	}
	return out, nil
}
