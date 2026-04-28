package validate

import (
	"math"
	"regexp"
	"testing"
)

func TestStringSchema(t *testing.T) {
	schema := String().Email().Min(3).Max(50)

	if !schema.Validate("a@b.com").Ok {
		t.Fatal("expected valid email")
	}
	if schema.Validate("x").Ok {
		t.Fatal("expected short string to fail")
	}
	if schema.Validate("not-an-email").Ok {
		t.Fatal("expected invalid email to fail")
	}

	letters := String().Regex(regexp.MustCompile(`^[a-z]+$`))
	if !letters.Validate("abc").Ok {
		t.Fatal("expected regex match")
	}
	if letters.Validate("abc123").Ok {
		t.Fatal("expected regex mismatch")
	}
}

func TestNumberSchema(t *testing.T) {
	ints := Number().Int()
	for _, value := range []any{10, 0, -5} {
		if !ints.Validate(value).Ok {
			t.Fatalf("expected %v to be integer", value)
		}
	}
	for _, value := range []any{10.1, -3.14, "10", math.NaN(), true, nil, math.Inf(1), math.Inf(-1)} {
		if ints.Validate(value).Ok {
			t.Fatalf("expected %v to fail integer validation", value)
		}
	}

	float := Number().Float()
	for _, value := range []any{10, 10.5, -3.14, 0, math.Inf(1), math.Inf(-1)} {
		if !float.Validate(value).Ok {
			t.Fatalf("expected %v to be a number", value)
		}
	}
	for _, value := range []any{"10", math.NaN()} {
		if float.Validate(value).Ok {
			t.Fatalf("expected %v to fail number validation", value)
		}
	}

	positiveInt := Number().Int().Positive()
	if !positiveInt.Validate(1).Ok {
		t.Fatal("expected positive int")
	}
	if positiveInt.Validate(0).Ok || positiveInt.Validate(-1).Ok || positiveInt.Validate(1.2).Ok {
		t.Fatal("expected non-positive or non-int values to fail")
	}

	bounded := Number().Min(10).Max(20)
	if !bounded.Validate(10).Ok {
		t.Fatal("expected minimum boundary to pass")
	}
	if bounded.Validate(21).Ok {
		t.Fatal("expected max violation")
	}
}

func TestBooleanSchema(t *testing.T) {
	schema := Boolean()

	if schema.Validate("").Ok {
		t.Fatal("expected string to fail boolean validation")
	}
	if !schema.Validate(false).Ok || !schema.Validate(true).Ok {
		t.Fatal("expected booleans to pass")
	}
	if schema.Validate(1).Ok || schema.Validate(nil).Ok {
		t.Fatal("expected non-booleans to fail")
	}

	openapi := schema.ToOpenAPI()
	if openapi["type"] != "boolean" {
		t.Fatalf("unexpected OpenAPI schema: %#v", openapi)
	}
}

func TestObjectSchema(t *testing.T) {
	user := Object(Shape{
		"email":  String().Email(),
		"age":    Number().Int().Positive().Optional(),
		"tags":   Array[string](String().Min(1)).Optional(),
		"active": Boolean(),
	})

	ok := user.Validate(map[string]any{"email": "a@b.com", "active": true})
	if !ok.Ok {
		t.Fatalf("expected user to validate, got %#v", ok.Errors)
	}
	if ok.Value["email"] != "a@b.com" {
		t.Fatalf("unexpected email: %#v", ok.Value["email"])
	}
	if ok.Value["age"] != nil {
		t.Fatalf("expected missing optional age to become nil, got %#v", ok.Value["age"])
	}

	if user.Validate(map[string]any{"email": "nope", "active": true}).Ok {
		t.Fatal("expected invalid email to fail")
	}

	strict := Object(Shape{"a": String()}).Strict()
	result := strict.Validate(map[string]any{"a": "x", "extra": 1})
	if result.Ok {
		t.Fatal("expected strict object to reject unknown key")
	}
}

func TestArraySchema(t *testing.T) {
	schema := Array[float64](Number().Int()).Min(2).Max(3)

	if !schema.Validate([]any{1, 2}).Ok {
		t.Fatal("expected array to validate")
	}
	if schema.Validate([]any{1}).Ok {
		t.Fatal("expected minItems failure")
	}
	if schema.Validate([]any{1, 2, 3, 4}).Ok {
		t.Fatal("expected maxItems failure")
	}
	if schema.Validate([]any{1, 2.2}).Ok {
		t.Fatal("expected item validation failure")
	}
}

func TestNullableAndOptional(t *testing.T) {
	nullable := Boolean().Nullable()
	if !nullable.Validate(nil).Ok {
		t.Fatal("expected null boolean to validate")
	}
	if nullable.Validate("").Ok {
		t.Fatal("expected invalid nullable boolean to fail")
	}

	optionalNullable := Optional[*bool](nullable)
	result := Object(Shape{"flag": optionalNullable}).Validate(map[string]any{})
	if !result.Ok {
		t.Fatalf("expected missing optional nullable field to validate: %#v", result.Errors)
	}
}

func TestOpenAPI(t *testing.T) {
	createUser := Object(Shape{
		"email": String().Email(),
		"name":  String().Min(1),
	})
	ok := Object(Shape{"ok": Boolean()})

	doc := OpenAPI(OpenAPIOptions{
		Title:   "Test",
		Version: "1.0.0",
		Routes: []Route{
			OpenAPIRoute(Route{
				Method:      POST,
				Path:        "/users",
				RequestBody: createUser,
				Responses:   map[int]AnySchema{200: ok},
			}),
		},
	})

	if doc["openapi"] != "3.1.0" {
		t.Fatalf("unexpected OpenAPI version: %#v", doc["openapi"])
	}

	paths := doc["paths"].(map[string]any)
	post := paths["/users"].(map[string]any)["post"].(map[string]any)
	requestBody := post["requestBody"].(map[string]any)
	content := requestBody["content"].(map[string]any)
	requestSchema := content["application/json"].(map[string]any)["schema"].(map[string]any)
	if requestSchema["type"] != "object" {
		t.Fatalf("unexpected request schema: %#v", requestSchema)
	}

	responses := post["responses"].(map[string]any)
	responseContent := responses["200"].(map[string]any)["content"].(map[string]any)
	responseSchema := responseContent["application/json"].(map[string]any)["schema"].(map[string]any)
	if responseSchema["type"] != "object" {
		t.Fatalf("unexpected response schema: %#v", responseSchema)
	}
}

func TestValidateTargets(t *testing.T) {
	body := Object(Shape{"name": String().Min(2)})
	input := RequestData{
		Body:   map[string]any{"name": "A"},
		Query:  map[string]any{},
		Params: map[string]any{},
	}

	_, errors := Validate(Targets{Body: body}, input)
	if len(errors) == 0 {
		t.Fatal("expected validation error")
	}

	input.Body = map[string]any{"name": "Al"}
	validated, errors := Validate(Targets{Body: body}, input)
	if len(errors) != 0 {
		t.Fatalf("expected no validation errors, got %#v", errors)
	}
	if validated.Body.(map[string]any)["name"] != "Al" {
		t.Fatalf("unexpected validated body: %#v", validated.Body)
	}
}
