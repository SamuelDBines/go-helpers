package validate

type Targets struct {
	Body   AnySchema
	Query  AnySchema
	Params AnySchema
}

type RequestData struct {
	Body   any
	Query  any
	Params any
}

func Validate(targets Targets, input RequestData) (RequestData, []Issue) {
	out := input
	errors := make([]Issue, 0)

	if targets.Params != nil {
		result := targets.Params.ValidateAny(input.Params, "params")
		if result.Ok {
			out.Params = result.Value
		} else {
			errors = append(errors, result.Errors...)
		}
	}

	if targets.Query != nil {
		result := targets.Query.ValidateAny(input.Query, "query")
		if result.Ok {
			out.Query = result.Value
		} else {
			errors = append(errors, result.Errors...)
		}
	}

	if targets.Body != nil {
		result := targets.Body.ValidateAny(input.Body, "body")
		if result.Ok {
			out.Body = result.Value
		} else {
			errors = append(errors, result.Errors...)
		}
	}

	return out, errors
}
