package validate

import (
	"regexp"
	"strconv"
)

type stringCheck func(value string, path []any) *Issue

type StringSchema struct {
	meta   map[string]any
	checks []stringCheck
}

func String() *StringSchema {
	return &StringSchema{meta: map[string]any{"type": "string"}}
}

func (s *StringSchema) Kind() string {
	return "string"
}

func (s *StringSchema) Validate(input any, path ...any) Result[string] {
	value, ok := input.(string)
	if !ok {
		return Result[string]{
			Ok:     false,
			Errors: []Issue{typeIssue(path, "Expected string")},
		}
	}

	errors := make([]Issue, 0)
	for _, check := range s.checks {
		if issue := check(value, path); issue != nil {
			errors = append(errors, *issue)
		}
	}

	if len(errors) > 0 {
		return Result[string]{Ok: false, Errors: errors}
	}

	return Result[string]{Ok: true, Value: value}
}

func (s *StringSchema) ValidateAny(input any, path ...any) AnyResult {
	return anyResult(s.Validate(input, path...))
}

func (s *StringSchema) Min(n int) *StringSchema {
	s.meta["minLength"] = n
	s.checks = append(s.checks, func(value string, path []any) *Issue {
		if len(value) >= n {
			return nil
		}
		return &Issue{Path: clonePath(path), Code: "min", Message: "Min length " + strconv.Itoa(n)}
	})
	return s
}

func (s *StringSchema) Max(n int) *StringSchema {
	s.meta["maxLength"] = n
	s.checks = append(s.checks, func(value string, path []any) *Issue {
		if len(value) <= n {
			return nil
		}
		return &Issue{Path: clonePath(path), Code: "max", Message: "Max length " + strconv.Itoa(n)}
	})
	return s
}

func (s *StringSchema) Length(n int) *StringSchema {
	s.meta["minLength"] = n
	s.meta["maxLength"] = n
	s.checks = append(s.checks, func(value string, path []any) *Issue {
		if len(value) == n {
			return nil
		}
		return &Issue{Path: clonePath(path), Code: "length", Message: "Length " + strconv.Itoa(n)}
	})
	return s
}

func (s *StringSchema) Regex(re *regexp.Regexp) *StringSchema {
	s.meta["pattern"] = re.String()
	s.checks = append(s.checks, func(value string, path []any) *Issue {
		if re.MatchString(value) {
			return nil
		}
		return &Issue{Path: clonePath(path), Code: "pattern", Message: "Invalid format"}
	})
	return s
}

func (s *StringSchema) Email() *StringSchema {
	re := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	s.meta["format"] = "email"
	if _, ok := s.meta["pattern"]; !ok {
		s.meta["pattern"] = re.String()
	}
	s.checks = append(s.checks, func(value string, path []any) *Issue {
		if re.MatchString(value) {
			return nil
		}
		return &Issue{Path: clonePath(path), Code: "email", Message: "Invalid email"}
	})
	return s
}

func (s *StringSchema) UUID() *StringSchema {
	re := regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-5][0-9a-f]{3}-[089ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	s.meta["format"] = "uuid"
	if _, ok := s.meta["pattern"]; !ok {
		s.meta["pattern"] = re.String()
	}
	s.checks = append(s.checks, func(value string, path []any) *Issue {
		if re.MatchString(value) {
			return nil
		}
		return &Issue{Path: clonePath(path), Code: "uuid", Message: "Invalid uuid string"}
	})
	return s
}

func (s *StringSchema) Optional() *OptionalSchema[string] {
	return Optional[string](s)
}

func (s *StringSchema) Nullable() *NullableSchema[string] {
	return Nullable[string](s)
}

func (s *StringSchema) ToOpenAPI() map[string]any {
	return cloneOpenAPI(s.meta)
}

func (s *StringSchema) IsOptional() bool {
	return false
}
