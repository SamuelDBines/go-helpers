package validate

import "reflect"

type Shape map[string]AnySchema

type ObjectSchema struct {
	shape  Shape
	strict bool
}

func Object(shape Shape) *ObjectSchema {
	return &ObjectSchema{shape: shape}
}

func (s *ObjectSchema) Kind() string {
	return "object"
}

func (s *ObjectSchema) Validate(input any, path ...any) Result[map[string]any] {
	obj, ok := asStringMap(input)
	if !ok {
		return Result[map[string]any]{
			Ok:     false,
			Errors: []Issue{typeIssue(path, "Expected object")},
		}
	}

	out := make(map[string]any, len(s.shape))
	errors := make([]Issue, 0)

	if s.strict {
		for key := range obj {
			if _, ok := s.shape[key]; !ok {
				errors = append(errors, Issue{
					Path:    pathWith(path, key),
					Code:    "unknown",
					Message: "Unknown key",
				})
			}
		}
	}

	for key, schema := range s.shape {
		value, exists := obj[key]
		if !exists {
			value = missing
		}

		result := schema.ValidateAny(value, pathWith(path, key)...)
		if result.Ok {
			out[key] = result.Value
			continue
		}
		errors = append(errors, result.Errors...)
	}

	if len(errors) > 0 {
		return Result[map[string]any]{Ok: false, Errors: errors}
	}

	return Result[map[string]any]{Ok: true, Value: out}
}

func (s *ObjectSchema) ValidateAny(input any, path ...any) AnyResult {
	return anyResult(s.Validate(input, path...))
}

func (s *ObjectSchema) Strict() *ObjectSchema {
	s.strict = true
	return s
}

func (s *ObjectSchema) Optional() *OptionalSchema[map[string]any] {
	return Optional[map[string]any](s)
}

func (s *ObjectSchema) Nullable() *NullableSchema[map[string]any] {
	return Nullable[map[string]any](s)
}

func (s *ObjectSchema) ToOpenAPI() map[string]any {
	properties := make(map[string]any, len(s.shape))
	required := make([]string, 0, len(s.shape))

	for key, schema := range s.shape {
		properties[key] = schema.ToOpenAPI()
		if !schema.IsOptional() {
			required = append(required, key)
		}
	}

	out := map[string]any{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		out["required"] = required
	}
	if s.strict {
		out["additionalProperties"] = false
	}

	return out
}

func (s *ObjectSchema) IsOptional() bool {
	return false
}

func asStringMap(input any) (map[string]any, bool) {
	if input == nil {
		return nil, false
	}

	if obj, ok := input.(map[string]any); ok {
		return obj, true
	}

	value := reflect.ValueOf(input)
	if value.Kind() != reflect.Map || value.Type().Key().Kind() != reflect.String {
		return nil, false
	}

	out := make(map[string]any, value.Len())
	for _, key := range value.MapKeys() {
		out[key.String()] = value.MapIndex(key).Interface()
	}
	return out, true
}
