package jsonfmt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJSONParser(t *testing.T) {
	parser := NewJSONParser()
	testStr := "Some of the test string\n```json\n{\"name\": \"abcd\",\n\"age\": 30\n}```"
	allTokens := make([]*Token, 0)

	for _, char := range testStr {
		allTokens = append(allTokens, parser.Parse(string(char))...)
	}

	allTokens = append(allTokens, parser.End()...)

	printTokenTree(allTokens, 0)
	assert.Equal(t, "{\"name\":\"abcd\",\"age\":30}", TokensToString(allTokens), "JSON String is not as expected")
}

func TestNewJSONParserWithNestedObject(t *testing.T) {
	parser := NewJSONParser()
	testStr := "Some of the test string\n```json\n{\"name\": \"abcd\",\n\"age\": 30\n,\"address\": {\"city\": \"New York\", \"zip\": 10001}}"
	allTokens := make([]*Token, 0)

	for _, char := range testStr {
		allTokens = append(allTokens, parser.Parse(string(char))...)
	}

	allTokens = append(allTokens, parser.End()...)

	printTokenTree(allTokens, 0)
	assert.Equal(t, "{\"name\":\"abcd\",\"age\":30,\"address\":{\"city\":\"New York\",\"zip\":10001}}", TokensToString(allTokens), "JSON String is not as expected")
}

func TestNewJSONParserAgainstJSONBodyMissingClose(t *testing.T) {
	parser := NewJSONParser()
	testStr := "Some of the test string\n```json\n{\"name\": \"abcd\",\n\"age\": 30\n```"
	allTokens := make([]*Token, 0)

	for _, char := range testStr {
		allTokens = append(allTokens, parser.Parse(string(char))...)
	}

	allTokens = append(allTokens, parser.End()...)

	printTokenTree(allTokens, 0)
	assert.Equal(t, "{\"name\":\"abcd\",\"age\":30}", TokensToString(allTokens), "JSON String is not as expected")
}

func TestNewJSONParserAgainstJSONBodyMixedQuotes(t *testing.T) {
	parser := NewJSONParser()
	testStr := "Some of the test string\n```json\n{\"name\": \"abcd\",\n'age': 30}"
	allTokens := make([]*Token, 0)

	for _, char := range testStr {
		allTokens = append(allTokens, parser.Parse(string(char))...)
	}

	allTokens = append(allTokens, parser.End()...)

	printTokenTree(allTokens, 0)
	assert.Equal(t, "{\"name\":\"abcd\",\"age\":30}", TokensToString(allTokens), "JSON String is not as expected")
}

func TestNewJSONParserAgainstJSONBodyArrayOfPrimitives(t *testing.T) {
	parser := NewJSONParser()
	testStr := "Some of the test string\n```json\n[1, \"abcd\", true, null]"
	allTokens := make([]*Token, 0)

	for _, char := range testStr {
		allTokens = append(allTokens, parser.Parse(string(char))...)
	}

	allTokens = append(allTokens, parser.End()...)

	printTokenTree(allTokens, 0)
	assert.Equal(t, "[1,\"abcd\",true,null]", TokensToString(allTokens), "JSON String is not as expected")
}

func TestNewJSONParserAgainstJSONBodyArrayOfObjects(t *testing.T) {
	parser := NewJSONParser()
	testStr := "Some of the test string\n```json\n[{\"name\": \"abcd\",\n\"age\": 30\n}, {\"name\": \"efgh\",\n\"age\": 40\n}]"
	allTokens := make([]*Token, 0)

	for _, char := range testStr {
		allTokens = append(allTokens, parser.Parse(string(char))...)
	}

	allTokens = append(allTokens, parser.End()...)

	printTokenTree(allTokens, 0)
	assert.Equal(t, "[{\"name\":\"abcd\",\"age\":30},{\"name\":\"efgh\",\"age\":40}]", TokensToString(allTokens), "JSON String is not as expected")
}

func TestNewJSONParserAgainstJSONBodyArraySwappingNested(t *testing.T) {
	parser := NewJSONParser()
	testStr := "Some of the test string\n```json\n[{\n\"elements\": [\n{ \"elements\": [\n{\"name\": \"abcd\"}]\n}\n]\n}]"
	allTokens := make([]*Token, 0)

	for _, char := range testStr {
		allTokens = append(allTokens, parser.Parse(string(char))...)
	}

	allTokens = append(allTokens, parser.End()...)

	printTokenTree(allTokens, 0)
	assert.Equal(t, "[{\"elements\":[{\"elements\":[{\"name\":\"abcd\"}]}]}]", TokensToString(allTokens), "JSON String is not as expected")
}

func TestNewJSONParserAgainstJSONBodyArrayMissingClose(t *testing.T) {
	parser := NewJSONParser()
	testStr := "Some of the test string\n```json\n[1000,{\"name\": \"abcd\",\n\"age\": 30"
	allTokens := make([]*Token, 0)

	for _, char := range testStr {
		allTokens = append(allTokens, parser.Parse(string(char))...)
	}

	allTokens = append(allTokens, parser.End()...)

	printTokenTree(allTokens, 0)
	assert.Equal(t, "[1000,{\"name\":\"abcd\",\"age\":30}]", TokensToString(allTokens), "JSON String is not as expected")
}

func printTokenTree(tokens []*Token, indent int) {
	for _, token := range tokens {
		indentStr := strings.Repeat("  ", indent)
		fmt.Printf("%s%s: %s\n", indentStr, tokenTypeToString(token.Type), token.Content) //nolint:forbidigo
		if len(token.Children) > 0 {
			printTokenTree(token.Children, indent+1)
		}
	}
}

func tokenTypeToString(tokenType TokenType) string {
	switch tokenType {
	case TokenTypeText:
		return "Text"
	case TokenTypeJSONObject:
		return "Object"
	case TokenTypeJSONArray:
		return "Array"
	case TokenTypeJSONField:
		return "Field"
	case TokenTypeJSONString:
		return "String"
	case TokenTypeJSONNumber:
		return "Number"
	case TokenTypeJSONBoolean:
		return "Boolean"
	case TokenTypeJSONNull:
		return "Null"
	default:
		return "Unknown"
	}
}
