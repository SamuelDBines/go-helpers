package validate

import "reflect"

type Issue struct {
	Path    []any  `json:"path"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Result[T any] struct {
	Ok     bool
	Value  T
	Errors []Issue
}

type AnyResult struct {
	Ok     bool
	Value  any
	Errors []Issue
}

type AnySchema interface {
	Kind() string
	ValidateAny(input any, path ...any) AnyResult
	ToOpenAPI() map[string]any
	IsOptional() bool
}

type Schema[T any] interface {
	AnySchema
	Validate(input any, path ...any) Result[T]
}

type HttpMethod string

const (
	GET    HttpMethod = "get"
	POST   HttpMethod = "post"
	PUT    HttpMethod = "put"
	PATCH  HttpMethod = "patch"
	DELETE HttpMethod = "delete"
)

type missingValue struct{}

var missing = missingValue{}

func clonePath(path []any) []any {
	out := make([]any, len(path))
	copy(out, path)
	return out
}

func pathWith(path []any, segment any) []any {
	out := make([]any, 0, len(path)+1)
	out = append(out, path...)
	out = append(out, segment)
	return out
}

func cloneOpenAPI(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func anyResult[T any](result Result[T]) AnyResult {
	if !result.Ok {
		return AnyResult{Ok: false, Errors: result.Errors}
	}
	return AnyResult{Ok: true, Value: normalizeAny(result.Value)}
}

func typeIssue(path []any, message string) Issue {
	return Issue{Path: clonePath(path), Code: "type", Message: message}
}

func normalizeAny(value any) any {
	if value == nil {
		return nil
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if v.IsNil() {
			return nil
		}
	}

	return value
}
