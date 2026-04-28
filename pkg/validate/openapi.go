package validate

type Route struct {
	Method      HttpMethod
	Path        string
	RequestBody AnySchema
	Params      AnySchema
	Responses   map[int]AnySchema
}

type OpenAPIOptions struct {
	Title   string
	Version string
	Routes  []Route
}

func OpenAPI(opts OpenAPIOptions) map[string]any {
	paths := map[string]any{}

	for _, route := range opts.Routes {
		pathItem, ok := paths[route.Path].(map[string]any)
		if !ok {
			pathItem = map[string]any{}
			paths[route.Path] = pathItem
		}

		operation := map[string]any{
			"responses": openAPIResponses(route.Responses),
		}

		if route.RequestBody != nil {
			operation["requestBody"] = map[string]any{
				"required": true,
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": route.RequestBody.ToOpenAPI(),
					},
				},
			}
		}

		pathItem[string(route.Method)] = operation
	}

	return map[string]any{
		"openapi": "3.1.0",
		"info": map[string]any{
			"title":   opts.Title,
			"version": opts.Version,
		},
		"paths": paths,
	}
}

func OpenAPIRoute(route Route) Route {
	return route
}

func openAPIResponses(responses map[int]AnySchema) map[string]any {
	out := make(map[string]any, len(responses))
	for code, schema := range responses {
		out[formatStatusCode(code)] = map[string]any{
			"description": "Response",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": schema.ToOpenAPI(),
				},
			},
		}
	}
	return out
}

func formatStatusCode(code int) string {
	if code == 0 {
		return "0"
	}

	negative := code < 0
	if negative {
		code = -code
	}

	buf := make([]byte, 0, 3)
	for code > 0 {
		buf = append(buf, byte('0'+code%10))
		code /= 10
	}
	if negative {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
