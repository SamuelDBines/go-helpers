package gomon

import (
	"fmt"
	"strings"
)

func parseSimpleYAMLConfig(data []byte) (map[string]Job, error) {
	lines := strings.Split(string(data), "\n")
	jobs := make(map[string]Job)
	var cur string
	var j Job
	var listField string

	flush := func() {
		if cur != "" {
			jobs[cur] = j
		}
	}

	for _, raw := range lines {
		line := stripYAMLComment(strings.TrimRight(raw, "\r"))
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := countIndent(line)
		trim := strings.TrimSpace(line)

		if indent == 0 && strings.HasSuffix(trim, ":") && !strings.HasPrefix(trim, "-") {
			flush()
			cur = strings.TrimSuffix(trim, ":")
			cur = strings.TrimSpace(cur)
			if cur == "" {
				return nil, fmt.Errorf("empty job name")
			}
			j = Job{}
			listField = ""
			continue
		}
		if cur == "" {
			return nil, fmt.Errorf("expected top-level job before %q", trim)
		}
		if indent == 2 && strings.HasSuffix(trim, ":") && !strings.HasPrefix(trim, "-") {
			key := strings.TrimSuffix(trim, ":")
			key = strings.TrimSpace(key)
			listField = key
			switch key {
			case "exec":
				j.Exec.Parts = nil
			case "ignore", "watch", "ext":
				switch key {
				case "ignore":
					j.Ignore = nil
				case "watch":
					j.Watch = nil
				case "ext":
					j.Ext = nil
				}
			case "parallel":
				listField = ""
			default:
				return nil, fmt.Errorf("job %q: unknown key %q", cur, key)
			}
			continue
		}
		if indent == 2 && strings.Contains(trim, ":") && !strings.HasPrefix(strings.TrimSpace(trim), "-") {
			k, v, ok := strings.Cut(trim, ":")
			if !ok {
				continue
			}
			k = strings.TrimSpace(k)
			v = strings.TrimSpace(v)
			switch k {
			case "parallel":
				listField = ""
				switch strings.ToLower(v) {
				case "true":
					j.Parallel = true
				case "false":
					j.Parallel = false
				default:
					return nil, fmt.Errorf("job %q: parallel must be true or false", cur)
				}
			case "exec":
				listField = ""
				if strings.HasPrefix(v, "[") {
					inner := strings.TrimSuffix(strings.TrimPrefix(v, "["), "]")
					parts, err := splitYAMLFlowStrings(inner)
					if err != nil {
						return nil, fmt.Errorf("job %q: exec: %w", cur, err)
					}
					j.Exec.Parts = parts
				} else {
					j.Exec.Parts = []string{v}
				}
			case "ignore", "watch", "ext":
				listField = k
				if v != "" && v != "|" {
					if strings.HasPrefix(v, "[") {
						inner := strings.TrimSuffix(strings.TrimPrefix(v, "["), "]")
						parts, err := splitYAMLFlowStrings(inner)
						if err != nil {
							return nil, fmt.Errorf("job %q: %s: %w", cur, k, err)
						}
						setListField(&j, k, parts)
						listField = ""
					} else {
						setListField(&j, k, []string{v})
						listField = ""
					}
				}
			default:
				return nil, fmt.Errorf("job %q: unknown key %q", cur, k)
			}
			continue
		}
		if indent >= 4 && strings.HasPrefix(trim, "-") {
			item := strings.TrimSpace(strings.TrimPrefix(trim, "-"))
			item = strings.Trim(item, `"'`)
			switch listField {
			case "exec":
				j.Exec.Parts = append(j.Exec.Parts, item)
			case "ignore":
				j.Ignore = append(j.Ignore, item)
			case "watch":
				j.Watch = append(j.Watch, item)
			case "ext":
				j.Ext = append(j.Ext, item)
			default:
				return nil, fmt.Errorf("job %q: list item outside of a list key", cur)
			}
			continue
		}
		return nil, fmt.Errorf("job %q: unexpected line %q", cur, trim)
	}
	flush()
	if len(jobs) == 0 {
		return nil, fmt.Errorf("no jobs found")
	}
	return jobs, nil
}

func setListField(j *Job, field string, parts []string) {
	switch field {
	case "ignore":
		j.Ignore = parts
	case "watch":
		j.Watch = parts
	case "ext":
		j.Ext = parts
	case "exec":
		j.Exec.Parts = parts
	}
}

func stripYAMLComment(line string) string {
	inQuote := false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '"' || c == '\'' {
			inQuote = !inQuote
		}
		if !inQuote && c == '#' {
			return line[:i]
		}
	}
	return line
}

func countIndent(line string) int {
	n := 0
	for _, r := range line {
		if r == ' ' {
			n++
			continue
		}
		if r == '\t' {
			n += 4
			continue
		}
		break
	}
	return n
}

func splitYAMLFlowStrings(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	var out []string
	for len(strings.TrimSpace(s)) > 0 {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, `"`) {
			end := strings.Index(s[1:], `"`)
			if end < 0 {
				return nil, fmt.Errorf("unterminated string in %q", s)
			}
			out = append(out, s[1:end+1])
			s = strings.TrimSpace(s[end+2:])
		} else {
			if idx := strings.Index(s, ","); idx >= 0 {
				out = append(out, strings.TrimSpace(s[:idx]))
				s = s[idx+1:]
				continue
			}
			out = append(out, strings.TrimSpace(s))
			break
		}
		if strings.HasPrefix(s, ",") {
			s = strings.TrimSpace(s[1:])
		}
	}
	return out, nil
}
