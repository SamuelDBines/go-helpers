package validate

type OptionalSchema[T any] struct {
	inner Schema[T]
}

func Optional[T any](inner Schema[T]) *OptionalSchema[T] {
	return &OptionalSchema[T]{inner: inner}
}

func (s *OptionalSchema[T]) Kind() string {
	return s.inner.Kind() + ".optional"
}

func (s *OptionalSchema[T]) Validate(input any, path ...any) Result[*T] {
	if _, ok := input.(missingValue); ok {
		return Result[*T]{Ok: true}
	}

	result := s.inner.Validate(input, path...)
	if !result.Ok {
		return Result[*T]{Ok: false, Errors: result.Errors}
	}

	return Result[*T]{Ok: true, Value: &result.Value}
}

func (s *OptionalSchema[T]) ValidateAny(input any, path ...any) AnyResult {
	return anyResult(s.Validate(input, path...))
}

func (s *OptionalSchema[T]) ToOpenAPI() map[string]any {
	return s.inner.ToOpenAPI()
}

func (s *OptionalSchema[T]) IsOptional() bool {
	return true
}

type NullableSchema[T any] struct {
	inner Schema[T]
}

func Nullable[T any](inner Schema[T]) *NullableSchema[T] {
	return &NullableSchema[T]{inner: inner}
}

func (s *NullableSchema[T]) Kind() string {
	return s.inner.Kind() + ".nullable"
}

func (s *NullableSchema[T]) Validate(input any, path ...any) Result[*T] {
	if input == nil {
		return Result[*T]{Ok: true}
	}

	result := s.inner.Validate(input, path...)
	if !result.Ok {
		return Result[*T]{Ok: false, Errors: result.Errors}
	}

	return Result[*T]{Ok: true, Value: &result.Value}
}

func (s *NullableSchema[T]) ValidateAny(input any, path ...any) AnyResult {
	return anyResult(s.Validate(input, path...))
}

func (s *NullableSchema[T]) ToOpenAPI() map[string]any {
	return map[string]any{
		"anyOf": []any{s.inner.ToOpenAPI(), map[string]any{"type": "null"}},
	}
}

func (s *NullableSchema[T]) IsOptional() bool {
	return s.inner.IsOptional()
}
