package validate

import (
	"reflect"
	"strconv"
)

type ArraySchema[Item any] struct {
	item     Schema[Item]
	minItems *int
	maxItems *int
}

func Array[Item any](item Schema[Item]) *ArraySchema[Item] {
	return &ArraySchema[Item]{item: item}
}

func (s *ArraySchema[Item]) Kind() string {
	return "array"
}

func (s *ArraySchema[Item]) Validate(input any, path ...any) Result[[]Item] {
	value := reflect.ValueOf(input)
	if !value.IsValid() || (value.Kind() != reflect.Slice && value.Kind() != reflect.Array) {
		return Result[[]Item]{
			Ok:     false,
			Errors: []Issue{typeIssue(path, "Expected array")},
		}
	}

	errors := make([]Issue, 0)
	length := value.Len()

	if s.minItems != nil && length < *s.minItems {
		errors = append(errors, Issue{
			Path:    clonePath(path),
			Code:    "minItems",
			Message: "Min items " + strconv.Itoa(*s.minItems),
		})
	}

	if s.maxItems != nil && length > *s.maxItems {
		errors = append(errors, Issue{
			Path:    clonePath(path),
			Code:    "maxItems",
			Message: "Max items " + strconv.Itoa(*s.maxItems),
		})
	}

	out := make([]Item, 0, length)
	for i := 0; i < length; i++ {
		result := s.item.Validate(value.Index(i).Interface(), pathWith(path, i)...)
		if result.Ok {
			out = append(out, result.Value)
			continue
		}
		errors = append(errors, result.Errors...)
	}

	if len(errors) > 0 {
		return Result[[]Item]{Ok: false, Errors: errors}
	}

	return Result[[]Item]{Ok: true, Value: out}
}

func (s *ArraySchema[Item]) ValidateAny(input any, path ...any) AnyResult {
	return anyResult(s.Validate(input, path...))
}

func (s *ArraySchema[Item]) Min(n int) *ArraySchema[Item] {
	s.minItems = &n
	return s
}

func (s *ArraySchema[Item]) Max(n int) *ArraySchema[Item] {
	s.maxItems = &n
	return s
}

func (s *ArraySchema[Item]) Length(n int) *ArraySchema[Item] {
	s.minItems = &n
	s.maxItems = &n
	return s
}

func (s *ArraySchema[Item]) Optional() *OptionalSchema[[]Item] {
	return Optional[[]Item](s)
}

func (s *ArraySchema[Item]) Nullable() *NullableSchema[[]Item] {
	return Nullable[[]Item](s)
}

func (s *ArraySchema[Item]) ToOpenAPI() map[string]any {
	out := map[string]any{
		"type":  "array",
		"items": s.item.ToOpenAPI(),
	}

	if s.minItems != nil {
		out["minItems"] = *s.minItems
	}
	if s.maxItems != nil {
		out["maxItems"] = *s.maxItems
	}

	return out
}

func (s *ArraySchema[Item]) IsOptional() bool {
	return false
}
