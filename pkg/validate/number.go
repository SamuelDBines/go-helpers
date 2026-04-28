package validate

import (
	"fmt"
	"math"
	"reflect"
)

type numberCheck func(value float64, path []any) *Issue

type NumberSchema struct {
	meta   map[string]any
	checks []numberCheck
}

func Number() *NumberSchema {
	return &NumberSchema{meta: map[string]any{"type": "number"}}
}

func (s *NumberSchema) Kind() string {
	return "number"
}

func (s *NumberSchema) Validate(input any, path ...any) Result[float64] {
	value, ok := asFloat64(input)
	if !ok || math.IsNaN(value) {
		return Result[float64]{
			Ok:     false,
			Errors: []Issue{typeIssue(path, "Expected number")},
		}
	}

	errors := make([]Issue, 0)
	for _, check := range s.checks {
		if issue := check(value, path); issue != nil {
			errors = append(errors, *issue)
		}
	}

	if len(errors) > 0 {
		return Result[float64]{Ok: false, Errors: errors}
	}

	return Result[float64]{Ok: true, Value: value}
}

func (s *NumberSchema) ValidateAny(input any, path ...any) AnyResult {
	return anyResult(s.Validate(input, path...))
}

func (s *NumberSchema) Min(n float64) *NumberSchema {
	s.meta["minimum"] = n
	s.checks = append(s.checks, func(value float64, path []any) *Issue {
		if value >= n {
			return nil
		}
		return &Issue{Path: clonePath(path), Code: "min", Message: "Min " + formatNumber(n)}
	})
	return s
}

func (s *NumberSchema) Max(n float64) *NumberSchema {
	s.meta["maximum"] = n
	s.checks = append(s.checks, func(value float64, path []any) *Issue {
		if value <= n {
			return nil
		}
		return &Issue{Path: clonePath(path), Code: "max", Message: "Max " + formatNumber(n)}
	})
	return s
}

func (s *NumberSchema) Positive() *NumberSchema {
	s.meta["exclusiveMinimum"] = 0
	s.checks = append(s.checks, func(value float64, path []any) *Issue {
		if value > 0 {
			return nil
		}
		return &Issue{Path: clonePath(path), Code: "positive", Message: "Must be > 0"}
	})
	return s
}

func (s *NumberSchema) Negative() *NumberSchema {
	s.meta["exclusiveMaximum"] = 0
	s.checks = append(s.checks, func(value float64, path []any) *Issue {
		if value < 0 {
			return nil
		}
		return &Issue{Path: clonePath(path), Code: "negative", Message: "Must be < 0"}
	})
	return s
}

func (s *NumberSchema) Int() *NumberSchema {
	s.meta["type"] = "integer"
	s.checks = append(s.checks, func(value float64, path []any) *Issue {
		if !math.IsInf(value, 0) && math.Trunc(value) == value {
			return nil
		}
		return &Issue{Path: clonePath(path), Code: "int", Message: "Must be integer"}
	})
	return s
}

func (s *NumberSchema) Float() *NumberSchema {
	return s
}

func (s *NumberSchema) Double() *NumberSchema {
	return s
}

func (s *NumberSchema) Optional() *OptionalSchema[float64] {
	return Optional[float64](s)
}

func (s *NumberSchema) Nullable() *NullableSchema[float64] {
	return Nullable[float64](s)
}

func (s *NumberSchema) ToOpenAPI() map[string]any {
	return cloneOpenAPI(s.meta)
}

func (s *NumberSchema) IsOptional() bool {
	return false
}

func asFloat64(input any) (float64, bool) {
	if input == nil {
		return 0, false
	}

	value := reflect.ValueOf(input)
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(value.Uint()), true
	case reflect.Float32, reflect.Float64:
		return value.Convert(reflect.TypeOf(float64(0))).Float(), true
	default:
		return 0, false
	}
}

func formatNumber(value float64) string {
	return fmt.Sprintf("%g", value)
}
