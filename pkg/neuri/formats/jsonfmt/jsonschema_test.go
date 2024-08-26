package jsonfmt

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/samber/lo"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func toJSONString(v any) string {
	return string(lo.Must(json.Marshal(v)))
}

func createDistortions(str string, primitiveType bool) []string {
	distortions := []string{
		fmt.Sprintf("Sure, here's the JSON: %s", str),
		fmt.Sprintf("The JSON object is: \n```json\n%s\n```", str),
		fmt.Sprintf("Here's what you asked for:\n%s\nIs there anything else?", str),
		fmt.Sprintf("[%s]", str),
	}
	if !primitiveType {
		distortions = append(distortions, fmt.Sprintf(`{"result": %s}`, str))
	}

	return distortions
}

func createInvalidDistortions(str string, primitiveType bool) []string {
	distortions := []string{
		fmt.Sprintf("This is invalid: %s", str),
		fmt.Sprintf("```json\n%s\n```", str),
	}
	if !primitiveType {
		distortions = append(distortions, fmt.Sprintf(`{"invalid": %s}`, str))
	}

	return distortions
}

func TestBuildRegexFromSchemaString(t *testing.T) {
	t.Run("handles basic object schema", func(t *testing.T) {
		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "number"},
			},
		}

		schemaJSON, err := json.Marshal(schema)
		require.NoError(t, err)

		regexString, err := BuildRegexFromSchemaString(string(schemaJSON), "")
		require.NoError(t, err)

		regex := regexp.MustCompile(regexString)

		validJSON := `{"age":30,"name":"John"}`
		for _, distorted := range createDistortions(validJSON, false) {
			match := regex.FindString(distorted)
			assert.NotEmpty(t, match)
			assert.Contains(t, match, validJSON)
		}

		invalidJSON := `{"age":30,"name":123}`
		for _, distorted := range createInvalidDistortions(invalidJSON, false) {
			match := regex.FindString(distorted)
			assert.Empty(t, match)
		}
	})

	t.Run("handles arrays", func(t *testing.T) {
		schema := map[string]any{
			"type":  "array",
			"items": map[string]any{"type": "number"},
		}

		schemaJSON, err := json.Marshal(schema)
		require.NoError(t, err)

		regexString, err := BuildRegexFromSchemaString(string(schemaJSON), "")
		require.NoError(t, err)

		regex := regexp.MustCompile(regexString)

		validJSON := `[1,2,3]`
		for _, distorted := range createDistortions(validJSON, false) {
			match := regex.FindString(distorted)
			assert.NotEmpty(t, match)
			assert.Contains(t, match, validJSON)
		}

		invalidJSON := `["abcd","abcd","abcd"]`
		for _, distorted := range createInvalidDistortions(invalidJSON, false) {
			match := regex.FindString(distorted)
			assert.Empty(t, match)
		}
	})

	t.Run("handles string formats", func(t *testing.T) {
		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":   map[string]any{"type": "string", "format": "uuid"},
				"date": map[string]any{"type": "string", "format": "date-time"},
			},
		}

		schemaJSON, err := json.Marshal(schema)
		require.NoError(t, err)

		regexString, err := BuildRegexFromSchemaString(string(schemaJSON), "")
		require.NoError(t, err)

		regex := regexp.MustCompile(regexString)

		validJSON := `{"date":"2023-06-13T15:30:00Z","id":"123e4567-e89b-12d3-a456-426614174000"}`
		for _, distorted := range createDistortions(validJSON, false) {
			match := regex.FindString(distorted)
			assert.NotEmpty(t, match)
			assert.Contains(t, match, validJSON)
		}

		invalidJSON := `{"date":"2023-06-13","id":"not-a-uuid"}`
		for _, distorted := range createInvalidDistortions(invalidJSON, false) {
			match := regex.FindString(distorted)
			assert.Empty(t, match)
		}
	})

	t.Run("handles number constraints", func(t *testing.T) {
		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"integer": map[string]any{"type": "integer", "minimum": 0, "maximum": 100},
				"float":   map[string]any{"type": "number", "exclusiveMinimum": 0, "exclusiveMaximum": 1},
			},
		}

		schemaJSON, err := json.Marshal(schema)
		require.NoError(t, err)

		regexString, err := BuildRegexFromSchemaString(string(schemaJSON), "")
		require.NoError(t, err)

		regex := regexp.MustCompile(regexString)

		validJSON := `{"float":0.5,"integer":50}`
		for _, distorted := range createDistortions(validJSON, false) {
			match := regex.FindString(distorted)
			assert.NotEmpty(t, match)
			assert.Contains(t, match, validJSON)
		}

		invalidJSON := `{"float":"0.5","integer":"50"}`
		for _, distorted := range createInvalidDistortions(invalidJSON, false) {
			match := regex.FindString(distorted)
			assert.Empty(t, match)
		}
	})

	t.Run("handles required properties", func(t *testing.T) {
		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id":   map[string]any{"type": "number"},
				"name": map[string]any{"type": "string"},
			},
			"required": []string{"id"},
		}

		schemaJSON, err := json.Marshal(schema)
		require.NoError(t, err)

		regexString, err := BuildRegexFromSchemaString(string(schemaJSON), "")
		require.NoError(t, err)

		regex := regexp.MustCompile(regexString)

		validJSON := `{"id":1,"name":"John"}`
		for _, distorted := range createDistortions(validJSON, false) {
			match := regex.FindString(distorted)
			assert.NotEmpty(t, match)
			assert.Contains(t, match, validJSON)
		}

		invalidJSON := `{"name":"John"}`
		for _, distorted := range createInvalidDistortions(invalidJSON, false) {
			match := regex.FindString(distorted)
			assert.Empty(t, match)
		}
	})

	t.Run("handles string enums", func(t *testing.T) {
		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"color": map[string]any{"type": "string", "enum": []string{"red", "green", "blue"}},
			},
		}

		schemaJSON, err := json.Marshal(schema)
		require.NoError(t, err)

		regexString, err := BuildRegexFromSchemaString(string(schemaJSON), "")
		require.NoError(t, err)

		regex := regexp.MustCompile(regexString)

		validJSON := `{"color":"red"}`
		for _, distorted := range createDistortions(validJSON, false) {
			match := regex.FindString(distorted)
			assert.NotEmpty(t, match)
			assert.Contains(t, match, validJSON)
		}

		invalidJSON := `{"color":"yellow"}`
		for _, distorted := range createInvalidDistortions(invalidJSON, false) {
			match := regex.FindString(distorted)
			assert.Empty(t, match)
		}
	})

	t.Run("handles number enums", func(t *testing.T) {
		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"status": map[string]any{"type": "string", "enum": []int{1, 2, 3}},
			},
		}

		schemaJSON, err := json.Marshal(schema)
		require.NoError(t, err)

		regexString, err := BuildRegexFromSchemaString(string(schemaJSON), "")
		require.NoError(t, err)

		regex := regexp.MustCompile(regexString)

		validJSON := `{"status":1}`
		for _, distorted := range createDistortions(validJSON, false) {
			match := regex.FindString(distorted)
			assert.NotEmpty(t, match)
			assert.Contains(t, match, validJSON)
		}

		invalidJSON := `{"status":10}`
		for _, distorted := range createInvalidDistortions(invalidJSON, false) {
			match := regex.FindString(distorted)
			assert.Empty(t, match)
		}
	})

	t.Run("handles nested objects", func(t *testing.T) {
		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"person": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
						"age":  map[string]any{"type": "number"},
					},
				},
			},
		}

		schemaJSON, err := json.Marshal(schema)
		require.NoError(t, err)

		regexString, err := BuildRegexFromSchemaString(string(schemaJSON), "")
		require.NoError(t, err)

		regex := regexp.MustCompile(regexString)

		validJSON := `{"person":{"age":30,"name":"John"}}`
		for _, distorted := range createDistortions(validJSON, false) {
			match := regex.FindString(distorted)
			assert.NotEmpty(t, match)
			assert.Contains(t, match, validJSON)
		}

		invalidJSON := `{"person":{"age":"30","name":"John"}}`
		for _, distorted := range createInvalidDistortions(invalidJSON, false) {
			match := regex.FindString(distorted)
			assert.Empty(t, match)
		}
	})

	t.Run("handles oneOf", func(t *testing.T) {
		schema := map[string]any{
			"oneOf": []any{
				map[string]any{
					"type":       "object",
					"properties": map[string]any{"value": map[string]any{"type": "string"}},
				},
				map[string]any{
					"type":       "object",
					"properties": map[string]any{"value": map[string]any{"type": "number"}},
				},
			},
		}

		schemaJSON, err := json.Marshal(schema)
		require.NoError(t, err)

		regexString, err := BuildRegexFromSchemaString(string(schemaJSON), "")
		require.NoError(t, err)

		regex := regexp.MustCompile(regexString)

		validJSONs := []string{`{"value":"text"}`, `{"value":1}`}
		for _, validJSON := range validJSONs {
			for _, distorted := range createDistortions(validJSON, true) {
				match := regex.FindString(distorted)
				assert.NotEmpty(t, match)
				assert.Contains(t, match, validJSON)
			}
		}

		invalidJSON := `true`
		for _, distorted := range createInvalidDistortions(invalidJSON, true) {
			match := regex.FindString(distorted)
			assert.Empty(t, match)
		}
	})

	t.Run("handles allOf", func(t *testing.T) {
		schema := map[string]any{
			"allOf": []any{
				map[string]any{
					"type":       "object",
					"properties": map[string]any{"a": map[string]any{"type": "number"}},
					"required":   []string{"a"},
				},
				map[string]any{
					"type":       "object",
					"properties": map[string]any{"b": map[string]any{"type": "string"}},
					"required":   []string{"b"},
				},
			},
		}

		schemaJSON, err := json.Marshal(schema)
		require.NoError(t, err)

		regexString, err := BuildRegexFromSchemaString(string(schemaJSON), "")
		require.NoError(t, err)

		regex := regexp.MustCompile(regexString)

		validJSON := `{"a":1,"b":"text"}`
		for _, distorted := range createDistortions(validJSON, false) {
			match := regex.FindString(distorted)
			assert.NotEmpty(t, match)
			assert.Contains(t, match, validJSON)
		}

		invalidJSONs := []string{`{"a":1}`, `{"b":"text"}`, `{"a":"1","b":"text"}`}
		for _, invalidJSON := range invalidJSONs {
			for _, distorted := range createInvalidDistortions(invalidJSON, false) {
				match := regex.FindString(distorted)
				assert.Empty(t, match)
			}
		}
	})

	t.Run("handles anyOf", func(t *testing.T) {
		schema := map[string]any{
			"anyOf": []any{
				map[string]any{
					"type":       "object",
					"properties": map[string]any{"a": map[string]any{"type": "number"}},
					"required":   []string{"a"},
				},
				map[string]any{
					"type":       "object",
					"properties": map[string]any{"b": map[string]any{"type": "string"}},
					"required":   []string{"b"},
				},
			},
		}

		schemaJSON, err := json.Marshal(schema)
		require.NoError(t, err)

		regexString, err := BuildRegexFromSchemaString(string(schemaJSON), "")
		require.NoError(t, err)

		regex := regexp.MustCompile(regexString)

		validJSONs := []string{`{"a":1}`, `{"b":"text"}`, `{"a":1,"b":"text"}`}
		for _, validJSON := range validJSONs {
			for _, distorted := range createDistortions(validJSON, false) {
				match := regex.FindString(distorted)
				assert.NotEmpty(t, match)
				assert.Contains(t, match, validJSON)
			}
		}

		invalidJSONs := []string{`{"a":"1"}`, `{"b":2}`, `{}`}
		for _, invalidJSON := range invalidJSONs {
			for _, distorted := range createInvalidDistortions(invalidJSON, false) {
				match := regex.FindString(distorted)
				assert.Empty(t, match)
			}
		}
	})
}

func TestBuildRegexFromSchema(t *testing.T) {
	t.Run("handles basic object schema", func(t *testing.T) {
		schemaMap := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "number"},
			},
		}

		schema, err := jsonschema.CompileString("schema.json", toJSONString(schemaMap))
		require.NoError(t, err)

		regexString, err := BuildRegexFromSchema(schema, "")
		require.NoError(t, err)

		regex := regexp.MustCompile(regexString)

		validJSON := `{"age":30,"name":"John"}`
		for _, distorted := range createDistortions(validJSON, false) {
			match := regex.FindString(distorted)
			assert.NotEmpty(t, match)
			assert.Contains(t, match, validJSON)
		}

		invalidJSON := `{"age":30,"name":123}`
		for _, distorted := range createInvalidDistortions(invalidJSON, false) {
			match := regex.FindString(distorted)
			assert.Empty(t, match)
		}
	})
}

func TestExtractBySchema(t *testing.T) {
	t.Run("extracts object properties", func(t *testing.T) {
		schemaMap := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "number"},
			},
		}

		schema, err := jsonschema.CompileString("schema.json", toJSONString(schemaMap))
		require.NoError(t, err)

		input := `Sure, here's the JSON: {"age":30,"name":"John"}`
		extracted, err := ExtractBySchema(schema, input)
		require.NoError(t, err)
		assert.Equal(t, `{"age":30,"name":"John"}`, extracted)
	})
}

func TestExtractBySchemaWithParser(t *testing.T) {
	t.Run("extracts object properties", func(t *testing.T) {
		schemaMap := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "number"},
			},
		}

		schema, err := jsonschema.CompileString("schema.json", toJSONString(schemaMap))
		require.NoError(t, err)

		input := `Sure, here's the JSON: {"name":"John","age":30}`
		extracted, err := ExtractBySchemaWithParser(schema, input)
		require.NoError(t, err)
		assert.Equal(t, `{"age":30,"name":"John"}`, extracted)
	})
}

func TestExtractObjectBySchema(t *testing.T) {
	t.Run("extracts object properties", func(t *testing.T) {
		schemaMap := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
				"age":  map[string]any{"type": "number"},
			},
		}

		schema, err := jsonschema.CompileString("schema.json", toJSONString(schemaMap))
		require.NoError(t, err)

		input := "Sure, here's the JSON: ```json{\"age\":30,\"name\":\"John\"}```"

		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		extracted, err := ExtractStructBySchema[Person](schema, input)
		require.NoError(t, err)
		assert.Equal(t, Person{Name: "John", Age: 30}, *extracted)
	})
}
