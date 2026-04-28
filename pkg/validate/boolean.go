package validate

type BooleanSchema struct{}

func Boolean() *BooleanSchema {
	return &BooleanSchema{}
}

func (s *BooleanSchema) Kind() string {
	return "boolean"
}

func (s *BooleanSchema) Validate(input any, path ...any) Result[bool] {
	value, ok := input.(bool)
	if !ok {
		return Result[bool]{
			Ok:     false,
			Errors: []Issue{typeIssue(path, "Expected boolean")},
		}
	}

	return Result[bool]{Ok: true, Value: value}
}

func (s *BooleanSchema) ValidateAny(input any, path ...any) AnyResult {
	return anyResult(s.Validate(input, path...))
}

func (s *BooleanSchema) Optional() *OptionalSchema[bool] {
	return Optional[bool](s)
}

func (s *BooleanSchema) Nullable() *NullableSchema[bool] {
	return Nullable[bool](s)
}

func (s *BooleanSchema) ToOpenAPI() map[string]any {
	return map[string]any{"type": "boolean"}
}

func (s *BooleanSchema) IsOptional() bool {
	return false
}
