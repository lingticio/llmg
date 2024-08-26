package jsonschema

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/samber/lo"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
)

const (
	JSONSchemaRegExpStringInner = `[^"\\]*(?:\\.[^"\\]*)*`
	JSONSchemaRegExpString      = `"` + JSONSchemaRegExpStringInner + `"`
	JSONSchemaRegExpInteger     = `(-)?(0|[1-9][0-9]*)`
	JSONSchemaRegExpNumber      = JSONSchemaRegExpInteger + `(\.[0-9]+)?([eE][+-]?[0-9]+)?`
	JSONSchemaRegExpBoolean     = `(true|false)`
	JSONSchemaRegExpNull        = `null`
	JSONSchemaRegExpWhitespace  = `\s*`
	JSONSchemaRegExpDateTime    = `"(-?(?:[1-9][0-9]*)?[0-9]{4})-(1[0-2]|0[1-9])-(3[01]|0[1-9]|[12][0-9])T(2[0-3]|[01][0-9]):([0-5][0-9]):([0-5][0-9])(\.[0-9]{3})?(Z)?"`
	JSONSchemaRegExpDate        = `"(?:\d{4})-(?:0[1-9]|1[0-2])-(?:0[1-9]|[1-2][0-9]|3[0-1])"`
	JSONSchemaRegExpTime        = `"(2[0-3]|[01][0-9]):([0-5][0-9]):([0-5][0-9])(\\.[0-9]+)?(Z)?"`
	JSONSchemaRegExpUUID        = `"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"`
)

var TypeToRegexMap = map[string]string{
	"string":  JSONSchemaRegExpString,
	"integer": JSONSchemaRegExpInteger,
	"number":  JSONSchemaRegExpNumber,
	"boolean": JSONSchemaRegExpBoolean,
	"null":    JSONSchemaRegExpNull,
}

var FormatToRegexMap = map[string]string{
	"uuid":      JSONSchemaRegExpUUID,
	"date-time": JSONSchemaRegExpDateTime,
	"date":      JSONSchemaRegExpDate,
	"time":      JSONSchemaRegExpTime,
}

func handleEmptySchema(whitespacePattern string, rootSchema *jsonschema.Schema) (string, error) {
	types := []string{"boolean", "null", "number", "integer", "string", "array", "object"}
	regExps := make([]string, len(types))

	for i, t := range types {
		schema := &jsonschema.Schema{Types: []string{t}}

		regex, err := ToRegex(schema, whitespacePattern, rootSchema)
		if err != nil {
			return "", err
		}

		regExps[i] = fmt.Sprintf("(%s)", regex)
	}

	return strings.Join(regExps, "|"), nil
}

func handleProperties(properties map[string]*jsonschema.Schema, instance *jsonschema.Schema, whitespacePattern string, rootSchema *jsonschema.Schema) (string, error) {
	if len(properties) == 0 {
		return `\{\s*\}`, nil
	}

	// Get all property names and sort them
	propertyNames := make([]string, 0, len(properties))
	for name := range properties {
		propertyNames = append(propertyNames, name)
	}

	sort.Strings(propertyNames)

	propertyRegExps := make([]string, 0, len(propertyNames))
	isRequired := make(map[string]bool)

	for _, rp := range instance.Required {
		isRequired[rp] = true
	}

	for _, name := range propertyNames {
		value := properties[name]
		subRegex := fmt.Sprintf(`%s"%s"%s:%s`, whitespacePattern, regexp.QuoteMeta(name), whitespacePattern, whitespacePattern)

		valueRegex, err := ToRegex(value, whitespacePattern, rootSchema)
		if err != nil {
			return "", err
		}

		subRegex += valueRegex

		if isRequired[name] {
			propertyRegExps = append(propertyRegExps, subRegex)
		} else {
			propertyRegExps = append(propertyRegExps, fmt.Sprintf("(%s)?", subRegex))
		}
	}

	propertiesRegex := strings.Join(propertyRegExps, fmt.Sprintf(`%s,?%s`, whitespacePattern, whitespacePattern))
	objectRegex := fmt.Sprintf(`\{%s%s%s\}`, whitespacePattern, propertiesRegex, whitespacePattern)

	if instance.AdditionalProperties != nil {
		if additionalProps, ok := instance.AdditionalProperties.(bool); ok && additionalProps {
			// If additional properties are allowed, add a pattern for them
			objectRegex = fmt.Sprintf(`\{%s(%s%s,?%s)*(%s)?%s\}`,
				whitespacePattern,
				propertiesRegex,
				whitespacePattern,
				whitespacePattern,
				`"[^"]+"\s*:\s*[^,}]+`,
				whitespacePattern)
		}
	}

	return objectRegex, nil
}

func handleOneOf(oneOf []*jsonschema.Schema, whitespacePattern string, rootSchema *jsonschema.Schema) (string, error) {
	subRegExps := make([]string, len(oneOf))

	for i, schema := range oneOf {
		subRegex, err := ToRegex(schema, whitespacePattern, rootSchema)
		if err != nil {
			return "", err
		}
		subRegExps[i] = fmt.Sprintf("(?:%s)", subRegex)
	}

	return fmt.Sprintf("(%s)", strings.Join(subRegExps, "|")), nil
}

func handleAllOf(allOf []*jsonschema.Schema, whitespacePattern string, rootSchema *jsonschema.Schema) (string, error) {
	mergedProperties := make(map[string]*jsonschema.Schema)
	requiredProperties := make([]string, 0)

	for _, schema := range allOf {
		for propName, propSchema := range schema.Properties {
			mergedProperties[propName] = propSchema
		}
		requiredProperties = append(requiredProperties, schema.Required...)
	}

	regex := "\\{"
	propertyRegExps := make([]string, 0)

	for key, value := range mergedProperties {
		propertyRegex, err := ToRegex(value, whitespacePattern, rootSchema)
		if err != nil {
			return "", err
		}

		propPattern := fmt.Sprintf(`%s"%s"%s:%s%s`, whitespacePattern, key, whitespacePattern, whitespacePattern, propertyRegex)
		if !lo.Contains(requiredProperties, key) {
			propPattern = fmt.Sprintf("(%s)?", propPattern)
		}

		propertyRegExps = append(propertyRegExps, propPattern)
	}

	regex += strings.Join(propertyRegExps, fmt.Sprintf("%s,", whitespacePattern))
	regex += fmt.Sprintf("%s\\}", whitespacePattern)

	return regex, nil
}

func handleAnyOf(anyOf []*jsonschema.Schema, whitespacePattern string, rootSchema *jsonschema.Schema) (string, error) {
	subRegExps := make([]string, len(anyOf))
	var err error

	for i, schema := range anyOf {
		subRegExps[i], err = ToRegex(schema, whitespacePattern, rootSchema)
		if err != nil {
			return "", err
		}
	}

	// Generate all possible combinations
	combinations := []string{}

	for i := 1; i <= len(subRegExps); i++ {
		combos := getCombinations(subRegExps, i)
		for _, combo := range combos {
			combinations = append(combinations, combineSchemas(combo, whitespacePattern))
		}
	}

	return fmt.Sprintf("(%s)", strings.Join(combinations, "|")), nil
}

func combineSchemas(schemas []string, whitespacePattern string) string {
	// Remove outer curly braces from each schema
	for i, schema := range schemas {
		schemas[i] = strings.TrimPrefix(strings.TrimSuffix(schema, `\s*\}`), `\{\s*`)
	}

	// Join schemas with optional comma and whitespace
	combined := strings.Join(schemas, fmt.Sprintf(`%s,?%s`, whitespacePattern, whitespacePattern))

	// Add back the outer curly braces
	return fmt.Sprintf(`\{%s%s%s\}`, whitespacePattern, combined, whitespacePattern)
}

func getCombinations(arr []string, k int) [][]string {
	if k == 1 {
		result := make([][]string, len(arr))
		for i, v := range arr {
			result[i] = []string{v}
		}

		return result
	}

	result := [][]string{}

	for i := 0; i <= len(arr)-k; i++ {
		subCombos := getCombinations(arr[i+1:], k-1)
		for _, subCombo := range subCombos {
			result = append(result, append([]string{arr[i]}, subCombo...))
		}
	}

	return result
}

func handlePrefixItems(prefixItems []*jsonschema.Schema, instance *jsonschema.Schema, whitespacePattern string, rootSchema *jsonschema.Schema) (string, error) {
	elementPatterns := make([]string, len(prefixItems))

	for i, item := range prefixItems {
		pattern, err := ToRegex(item, whitespacePattern, rootSchema)
		if err != nil {
			return "", err
		}
		elementPatterns[i] = pattern
	}

	commaSplitPattern := fmt.Sprintf("%s,%s", whitespacePattern, whitespacePattern)
	tupleInner := strings.Join(elementPatterns, commaSplitPattern)
	regex := fmt.Sprintf("\\[%s%s", whitespacePattern, tupleInner)

	if items, ok := instance.Items.(*jsonschema.Schema); ok {
		additionalItemsRegex, err := ToRegex(items, whitespacePattern, rootSchema)
		if err != nil {
			return "", err
		}
		regex += fmt.Sprintf("(%s%s)*", commaSplitPattern, additionalItemsRegex)
	}

	regex += fmt.Sprintf("%s\\]", whitespacePattern)

	return regex, nil
}

func handleEnum(enum []interface{}) (string, error) {
	choices := make([]string, len(enum))

	for i, choice := range enum {
		var stringified string

		switch v := choice.(type) {
		case string:
			// For strings, use JSON marshaling to ensure proper quoting
			bytes, err := json.Marshal(v)
			if err != nil {
				return "", fmt.Errorf("failed to marshal string enum value: %w", err)
			}
			stringified = string(bytes)
		case float64, int, int64, float32:
			// For numbers, use fmt.Sprintf without quotes
			stringified = fmt.Sprintf("%v", v)
		case bool:
			// For booleans, use strings.ToLower to ensure "true" or "false"
			stringified = strings.ToLower(fmt.Sprintf("%v", v))
		case nil:
			stringified = "null"
		case json.Number:
			stringified = v.String()
		default:
			return "", fmt.Errorf("unsupported data type in enum: %T", v)
		}
		// Escape special regex characters
		choices[i] = regexp.QuoteMeta(stringified)
	}

	return fmt.Sprintf("(%s)", strings.Join(choices, "|")), nil
}

func handleConst(constValue interface{}) (string, error) {
	return regexp.QuoteMeta(fmt.Sprintf("%v", constValue)), nil
}

// func handleRef(ref string, rootSchema *jsonschema.Schema, whitespacePattern string) (string, error) {
// 	if strings.HasPrefix(ref, "#/") {
// 		refSchema, err := rootSchema.CompileRef(ref)
// 		if err != nil {
// 			return "", err
// 		}
// 		return toRegex(refSchema, whitespacePattern, rootSchema)
// 	}
// 	return "", fmt.Errorf("external references are not supported")
// }

func handleType(instance *jsonschema.Schema, whitespacePattern string, rootSchema *jsonschema.Schema) (string, error) {
	switch {
	case lo.Contains(instance.Types, "string"):
		return handleStringType(instance, whitespacePattern)
	case lo.Contains(instance.Types, "number"), lo.Contains(instance.Types, "integer"):
		return handleNumberType(instance)
	case lo.Contains(instance.Types, "array"):
		return handleArrayType(instance, whitespacePattern, rootSchema)
	case lo.Contains(instance.Types, "object"):
		return handleObjectType(instance, whitespacePattern, rootSchema)
	case lo.Contains(instance.Types, "boolean"):
		return TypeToRegexMap["boolean"], nil
	case lo.Contains(instance.Types, "null"):
		return TypeToRegexMap["null"], nil
	case len(instance.Types) > 1:
		return handleMultipleTypes(lo.Map(instance.Types, func(item string, _ int) any {
			return item
		}), whitespacePattern, rootSchema)
	default:
		return "", fmt.Errorf("invalid type specification")
	}
}

func handleStringType(instance *jsonschema.Schema, _ string) (string, error) {
	if instance.MaxLength > 0 {
		minLength := 0
		if instance.MinLength > 0 {
			minLength = instance.MinLength
		}

		return fmt.Sprintf(`"%s{%d,%d}"`, JSONSchemaRegExpStringInner, minLength, instance.MaxLength), nil
	} else if instance.MinLength > 0 {
		return fmt.Sprintf(`"%s{%d,}"`, JSONSchemaRegExpStringInner, instance.MinLength), nil
	} else if instance.Pattern != nil {
		pattern := instance.Pattern.String()
		if len(pattern) >= 2 && pattern[0] == '^' && pattern[len(pattern)-1] == '$' {
			return fmt.Sprintf(`("%s")`, pattern[1:len(pattern)-1]), nil
		}

		return fmt.Sprintf(`("%s")`, pattern), nil
	} else if instance.Format != "" {
		if regex, ok := FormatToRegexMap[instance.Format]; ok {
			return regex, nil
		}

		return "", fmt.Errorf("format %s is not supported", instance.Format)
	}

	// Default case: any string
	return JSONSchemaRegExpString, nil
}
func handleNumberType(instance *jsonschema.Schema) (string, error) {
	if lo.Contains(instance.Types, "integer") {
		return TypeToRegexMap["integer"], nil
	}

	return TypeToRegexMap["number"], nil
}

func handleArrayType(instance *jsonschema.Schema, whitespacePattern string, rootSchema *jsonschema.Schema) (string, error) {
	if instance.Items2020 == nil {
		return `\[\s*([^,\]]*\s*,?\s*)*\s*\]`, nil
	}

	itemsRegex, err := ToRegex(instance.Items2020, whitespacePattern, rootSchema)
	if err != nil {
		return "", err
	}

	if instance.MaxItems > 0 {
		minItems := 0
		if instance.MinItems != -1 {
			minItems = instance.MinItems
		}

		return fmt.Sprintf(`\[\s*(%s\s*,?\s*){%d,%d}\s*\]`, itemsRegex, minItems, instance.MaxItems), nil
	} else if instance.MinItems != -1 {
		return fmt.Sprintf(`\[\s*(%s\s*,?\s*){%d,}\s*\]`, itemsRegex, instance.MinItems), nil
	}

	return fmt.Sprintf(`\[\s*(%s\s*,?\s*){%d,}\s*\]`, itemsRegex, 0), nil
}

func handleObjectType(instance *jsonschema.Schema, whitespacePattern string, rootSchema *jsonschema.Schema) (string, error) {
	if len(instance.Properties) == 0 && instance.AdditionalProperties == nil {
		return `\{\s*\}`, nil
	}

	// Get all property names and sort them
	propertyNames := make([]string, 0, len(instance.Properties))
	for name := range instance.Properties {
		propertyNames = append(propertyNames, name)
	}

	sort.Strings(propertyNames)

	propertyRegexes := make([]string, 0, len(propertyNames))

	for _, name := range propertyNames {
		schema := instance.Properties[name]

		propertyRegex, err := ToRegex(schema, whitespacePattern, rootSchema)
		if err != nil {
			return "", err
		}

		propPattern := fmt.Sprintf(`"%s"\s*:\s*%s`, regexp.QuoteMeta(name), propertyRegex)
		if !lo.Contains(instance.Required, name) {
			propPattern = fmt.Sprintf(`(%s)?`, propPattern)
		}

		propertyRegexes = append(propertyRegexes, propPattern)
	}

	propertiesRegex := strings.Join(propertyRegexes, `\s*,?\s*`)
	objectRegex := fmt.Sprintf(`\{\s*%s\s*\}`, propertiesRegex)

	if instance.AdditionalProperties != nil {
		if additionalProps, ok := instance.AdditionalProperties.(bool); ok && additionalProps {
			// If additional properties are allowed, add a pattern for them
			objectRegex = fmt.Sprintf(`\{\s*(%s\s*,?\s*)*(%s)?\s*\}`, propertiesRegex, `"[^"]+"\s*:\s*[^,}]+`)
		}
	}

	return objectRegex, nil
}

func handleMultipleTypes(types []interface{}, whitespacePattern string, rootSchema *jsonschema.Schema) (string, error) {
	typesStr := lo.Map(types, func(t interface{}, _ int) string {
		str, _ := t.(string)
		return str
	})

	regExps := make([]string, 0)

	for _, t := range typesStr {
		schema := &jsonschema.Schema{Types: []string{t}}

		regex, err := ToRegex(schema, whitespacePattern, rootSchema)
		if err != nil {
			return "", err
		}

		regExps = append(regExps, regex)
	}

	return fmt.Sprintf("(%s)", strings.Join(regExps, "|")), nil
}

// Translate a JSON Schema instance into a regex that validates the schema.
// Many features of JSON schema are missing:
// - Handle `additionalProperties` keyword
// - Handle types defined as a list
// - Handle constraints on numbers
// - Handle special patterns: `date`, `uri`, etc.
//
// This does not support recursive definitions.
//
// Tweaked implementation from the original Python code to TypeScript of outlines
// https://github.com/outlines-dev/outlines/blob/8e94488d4ee3c5a29a919d0b9e19f7ea4170b1f4/outlines/fsm/json_schema.py#L142
func ToRegex(instance *jsonschema.Schema, whitespacePattern string, rootSchema *jsonschema.Schema) (string, error) {
	switch {
	case instance.Properties != nil:
		return handleProperties(instance.Properties, instance, whitespacePattern, rootSchema)
	case len(instance.AllOf) > 0:
		return handleAllOf(instance.AllOf, whitespacePattern, rootSchema)
	case len(instance.AnyOf) > 0:
		return handleAnyOf(instance.AnyOf, whitespacePattern, rootSchema)
	case len(instance.OneOf) > 0:
		return handleOneOf(instance.OneOf, whitespacePattern, rootSchema)
	case instance.PrefixItems != nil:
		return handlePrefixItems(instance.PrefixItems, instance, whitespacePattern, rootSchema)
	case instance.Enum != nil:
		return handleEnum(instance.Enum)
	case instance.Constant != nil:
		return handleConst(instance.Constant)
	case instance.Ref != nil:
		// return handleRef(instance.Ref, rootSchema, whitespacePattern)
	case len(instance.Types) > 0:
		return handleType(instance, whitespacePattern, rootSchema)
	case len(instance.Types) == 0:
		return handleEmptySchema(whitespacePattern, rootSchema)
	}

	return "", fmt.Errorf("unsupported schema type")
}

// Turn a plain text JSON schema into a regex that matches any JSON object that follows
// this schema.
//
// It works the same as {@link buildRegexFromSchema} but takes a string instead of an object.
//
// JSON Schema is a declarative language that allows to annotate JSON documents
// with types and descriptions. These schemas can be generated from any TypeScript
// JSON Schema tools, but not only limited to TypeScript, for example, an OpenAPI
// spec would do so and fit in the JSON schema's world, so does the Kubernetes CRD
// spec, as well as gRPC, tRPC, GraphQL, and many other API specs.
//
// And by ensuring that the generation respects the schema we ensure
// that the output can be parsed into these objects.
// This function parses the provided schema and builds a generation schedule which
// mixes deterministic generation (fixed strings), and sampling with constraints.
//
// References - [JSON Schema](https://json-schema.org/)
//
// Tweaked implementation from the original Python code to TypeScript of outlines
// https://github.com/outlines-dev/outlines/blob/8e94488d4ee3c5a29a919d0b9e19f7ea4170b1f4/outlines/fsm/json_schema.py#L44
func BuildRegexFromSchemaString(schema string, whitespacePattern string) (string, error) {
	s, err := jsonschema.CompileString("schema.json", schema)
	if err != nil {
		return "", err
	}

	regexp, err := BuildRegexFromSchema(s, whitespacePattern)
	if err != nil {
		return "", err
	}

	return regexp, nil
}

// Turn a JSON schema object into a regex that matches any JSON object that follows
// this schema.
// It works the same as {@link buildRegexFromSchemaString} but takes a JSON object instead of a string.
//
// JSON Schema is a declarative language that allows to annotate JSON documents
// with types and descriptions. These schemas can be generated from any TypeScript
// JSON Schema tools, but not only limited to TypeScript, for example, an OpenAPI
// spec would do so and fit in the JSON schema's world, so does the Kubernetes CRD
// spec, as well as gRPC, tRPC, GraphQL, and many other API specs.
//
// And by ensuring that the generation respects the schema we ensure
// that the output can be parsed into these objects.
// This function parses the provided schema and builds a generation schedule which
// mixes deterministic generation (fixed strings), and sampling with constraints.
//
// References - [JSON Schema](https://json-schema.org/)
//
// Tweaked implementation from the original Python code to TypeScript of outlines
// https://github.com/outlines-dev/outlines/blob/8e94488d4ee3c5a29a919d0b9e19f7ea4170b1f4/outlines/fsm/json_schema.py#L44
func BuildRegexFromSchema(schema *jsonschema.Schema, whitespacePattern string) (string, error) {
	if whitespacePattern == "" {
		whitespacePattern = JSONSchemaRegExpWhitespace
	}

	innerRegex, err := ToRegex(schema, whitespacePattern, schema)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s%s", JSONSchemaRegExpWhitespace, innerRegex, JSONSchemaRegExpWhitespace), nil
}

// Extracts the JSON object in plain text from a string that matches the provided schema.
//
// This function extracts the JSON object from a string that matches the provided schema.
// It uses the regular expression generated by BuildRegexFromSchema.
func ExtractBySchema(schema *jsonschema.Schema, extractFrom string) (string, error) {
	regex, err := BuildRegexFromSchema(schema, "")
	if err != nil {
		return "", err
	}

	re, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}

	match := re.FindString(extractFrom)
	if match == "" {
		return "", fmt.Errorf("no match found")
	}

	return strings.TrimSpace(match), nil
}

// Extracts a satisfied struct, or type from a string that matches the provided schema.
//
// This function extracts the JSON object from a string that matches the provided schema.
// It uses the regular expression generated by BuildRegexFromSchema.
func ExtractStructBySchema[T any](schema *jsonschema.Schema, extractFrom string) (*T, error) {
	extracted, err := ExtractBySchema(schema, extractFrom)
	if err != nil {
		return nil, err
	}

	var result T

	err = json.Unmarshal([]byte(extracted), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
