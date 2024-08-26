package jsonfmt

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenTypeText TokenType = iota
	TokenTypeJSONObject
	TokenTypeJSONArray
	TokenTypeJSONField
	TokenTypeJSONString
	TokenTypeJSONNumber
	TokenTypeJSONBoolean
	TokenTypeJSONNull
)

type Pos struct {
	Offset int
	Line   int
	Column int
}

type Token struct {
	Type     TokenType
	Content  string
	Pos      Pos
	Children []*Token
}

func (t *Token) stringifyObject() string {
	var parts []string

	for _, child := range t.Children {
		if child.Type == TokenTypeJSONField {
			parts = append(parts, child.String())
		}
	}

	return fmt.Sprintf("{%s}", strings.Join(parts, ","))
}

func (t *Token) stringifyArray() string {
	var parts []string

	for _, child := range t.Children {
		switch child.Type {
		case TokenTypeJSONObject, TokenTypeJSONArray:
			parts = append(parts, child.String())
		case TokenTypeJSONField:
			// This case handles improperly nested fields in arrays
			if len(child.Children) > 0 {
				parts = append(parts, child.Children[0].String())
			}
		case TokenTypeJSONBoolean:
			parts = append(parts, child.String())
		case TokenTypeJSONNull:
			parts = append(parts, child.String())
		case TokenTypeJSONString:
			parts = append(parts, child.String())
		case TokenTypeJSONNumber:
			parts = append(parts, child.String())
		case TokenTypeText:
			parts = append(parts, child.String())
		default:
			parts = append(parts, child.String())
		}
	}

	return fmt.Sprintf("[%s]", strings.Join(parts, ","))
}

func (t *Token) stringifyField() string {
	if len(t.Children) > 0 {
		childValue := t.Children[0].String()
		return fmt.Sprintf("\"%s\":%s", t.Content, childValue)
	}

	return fmt.Sprintf("\"%s\":null", t.Content)
}

func (t *Token) String() string {
	switch t.Type {
	case TokenTypeText:
		return t.Content
	case TokenTypeJSONObject:
		return t.stringifyObject()
	case TokenTypeJSONArray:
		return t.stringifyArray()
	case TokenTypeJSONField:
		return t.stringifyField()
	case TokenTypeJSONString:
		return t.Content
	case TokenTypeJSONNumber, TokenTypeJSONBoolean, TokenTypeJSONNull:
		return t.Content
	default:
		return ""
	}
}

type ParserState int

const (
	ParserStateText ParserState = iota
	ParserStateJSONStart
	ParserStateJSONString
	ParserStateJSONEscape
	ParserStateJSONFieldName
	ParserStateJSONFieldValue
	ParserStateJSONNumber
)

type JSONParser struct {
	buffer strings.Builder

	// States
	inSingleQuote bool
	insideJSON    bool
	state         ParserState
	stateStack    []ParserState
	depth         int
	tokenStart    Pos
	pos           Pos

	// Cursor
	currentToken     *Token
	currentContainer *Token

	// Tree structure
	tree []*Token
}

func NewJSONStreamParser() *JSONParser {
	return &JSONParser{
		state:      ParserStateText,
		stateStack: make([]ParserState, 0),
		tree:       make([]*Token, 0),
		pos:        Pos{Line: 1, Column: 1},
	}
}

func (p *JSONParser) Parse(chunk string) []*Token {
	for _, char := range chunk {
		p.processChar(char)
		p.updatePosition(char)
	}

	return p.getCompletedTokens()
}

func (p *JSONParser) End() []*Token {
	if p.insideJSON {
		p.autoCloseJSON()
	}

	p.flushBuffer()
	p.insideJSON = false

	return p.tree
}

func TokensToString(tokens []*Token) string {
	var result strings.Builder

	for _, token := range tokens {
		if token.Type == TokenTypeJSONObject || token.Type == TokenTypeJSONArray {
			result.WriteString(token.String())
		}
	}

	return result.String()
}

func (p *JSONParser) pushState(state ParserState) {
	p.stateStack = append(p.stateStack, state)
	p.state = state
}

func (p *JSONParser) popState() {
	if len(p.stateStack) > 1 {
		p.stateStack = p.stateStack[:len(p.stateStack)-1]
		p.state = p.stateStack[len(p.stateStack)-1]
	}
}

func (p *JSONParser) handleStateJSONNumber(char rune) {
	if unicode.IsDigit(char) || char == '.' || char == 'e' || char == 'E' || char == '+' || char == '-' {
		p.buffer.WriteRune(char)
	} else {
		p.completeCurrentToken()
		p.popState() // Return to previous state
		p.processChar(char)
	}
}

func (p *JSONParser) completeCurrentToken() {
	value := p.buffer.String()
	if value == "" {
		return // Don't create a token for an empty buffer
	}

	var tokenType TokenType

	switch {
	case p.state == ParserStateJSONString:
		tokenType = TokenTypeJSONString
		value = strings.Trim(value, "\"'") // Remove surrounding quotes
	case p.state == ParserStateJSONNumber:
		tokenType = TokenTypeJSONNumber
		value = trimTrailingNonNumber(value)
	case value == "true" || value == "false":
		tokenType = TokenTypeJSONBoolean
	case value == "null":
		tokenType = TokenTypeJSONNull
	default:
		tokenType = p.determineValueType(value)
	}

	valueToken := &Token{Type: tokenType, Content: value, Pos: p.tokenStart}

	if p.currentToken.Type == TokenTypeJSONField {
		p.currentToken.Children = append(p.currentToken.Children, valueToken)
	} else if p.currentContainer != nil {
		p.currentContainer.Children = append(p.currentContainer.Children, valueToken)
	}

	p.buffer.Reset()
	p.tokenStart = p.pos
}

func trimTrailingNonNumber(s string) string {
	for i := len(s) - 1; i >= 0; i-- {
		if unicode.IsDigit(rune(s[i])) || s[i] == '.' || s[i] == 'e' || s[i] == 'E' || s[i] == '+' || s[i] == '-' {
			return s[:i+1]
		}
	}

	return s
}

func (p *JSONParser) processChar(char rune) {
	switch p.state {
	case ParserStateText:
		p.handleStateText(char)
	case ParserStateJSONStart:
		p.handleStateJSONStart(char)
	case ParserStateJSONString:
		p.handleStateJSONString(char)
	case ParserStateJSONEscape:
		p.handleStateJSONEscape(char)
	case ParserStateJSONFieldName:
		p.handleStateJSONFieldName(char)
	case ParserStateJSONFieldValue:
		p.handleStateJSONFieldValue(char)
	case ParserStateJSONNumber:
		p.handleStateJSONNumber(char)
	}
}

func (p *JSONParser) handleStateText(char rune) {
	if char == '{' || char == '[' {
		p.flushBuffer()
		p.startNewJSONToken(char)
		p.insideJSON = true
		p.pushState(ParserStateJSONStart)
	} else {
		if p.buffer.Len() == 0 {
			p.tokenStart = p.pos
		}

		p.buffer.WriteRune(char)
	}
}

func (p *JSONParser) handleStateJSONStart(char rune) {
	switch {
	case char == '"' || char == '\'':
		p.pushState(ParserStateJSONString)
		p.inSingleQuote = (char == '\'')
		p.buffer.WriteRune(char)
	case char == '}' || char == ']':
		p.completeCurrentToken()
		p.depth--

		if p.depth == 0 {
			p.insideJSON = false
			p.popState() // Should return to StateText
			p.flushBuffer()
			p.currentContainer = nil
		} else {
			p.popState() // Return to previous state within JSON
			p.currentContainer = p.findParentContainer()
		}
	case char == '{' || char == '[':
		p.completeCurrentToken()
		p.startNewJSONToken(char)
	case char == ':':
		p.startNewJSONField()
		p.pushState(ParserStateJSONFieldValue)
	case char == ',':
		p.completeCurrentToken()
	case unicode.IsDigit(char) || char == '-':
		p.pushState(ParserStateJSONNumber)
		p.buffer.WriteRune(char)
	case char == 't' || char == 'f' || char == 'n':
		p.pushState(ParserStateJSONFieldValue)
		p.buffer.WriteRune(char)
	default:
		if !unicode.IsSpace(char) {
			if p.currentToken.Type == TokenTypeJSONArray {
				p.pushState(ParserStateJSONFieldValue)
			} else {
				p.pushState(ParserStateJSONFieldName)
			}

			p.buffer.WriteRune(char)
		}
	}
}

func (p *JSONParser) handleStateJSONString(char rune) {
	p.buffer.WriteRune(char)
	if char == '\\' {
		p.pushState(ParserStateJSONEscape)
	} else if (char == '"' && !p.inSingleQuote) || (char == '\'' && p.inSingleQuote) {
		p.popState()
		p.inSingleQuote = false
	}
}

func (p *JSONParser) handleStateJSONEscape(char rune) {
	p.buffer.WriteRune(char)
	p.popState() // Return to StateJSONString
}

func (p *JSONParser) handleStateJSONFieldName(char rune) {
	if char == ':' {
		p.startNewJSONField()
		p.popState()
		p.pushState(ParserStateJSONFieldValue)
	} else {
		p.buffer.WriteRune(char)
	}
}

func (p *JSONParser) handleStateJSONFieldValue(char rune) {
	switch {
	case char == ',' || char == '}' || char == ']':
		p.completeCurrentToken()
		p.popState() // Return to StateJSONStart
		if char == '}' || char == ']' {
			p.processChar(char)
		}
	case char == '{' || char == '[':
		p.startNewJSONToken(char)
	case char == '"' || char == '\'':
		p.pushState(ParserStateJSONString)
		p.inSingleQuote = (char == '\'')
		p.buffer.WriteRune(char)
	case unicode.IsDigit(char) || char == '-':
		p.popState()
		p.pushState(ParserStateJSONNumber)
		p.buffer.WriteRune(char)
	case char == 't' || char == 'f' || char == 'n':
		p.buffer.WriteRune(char)
	default:
		if !unicode.IsSpace(char) {
			p.buffer.WriteRune(char)
		}
	}
}

func (p *JSONParser) determineValueType(value string) TokenType {
	switch {
	case value == "true" || value == "false":
		return TokenTypeJSONBoolean
	case value == "null":
		return TokenTypeJSONNull
	case len(value) > 0 && (unicode.IsDigit(rune(value[0])) || value[0] == '-'):
		return TokenTypeJSONNumber
	default:
		return TokenTypeJSONString
	}
}

func (p *JSONParser) startNewJSONToken(char rune) {
	var tokenType TokenType
	if char == '{' {
		tokenType = TokenTypeJSONObject
	} else {
		tokenType = TokenTypeJSONArray
	}

	newToken := &Token{Type: tokenType, Pos: p.pos, Children: make([]*Token, 0)}
	if p.currentContainer != nil {
		p.currentContainer.Children = append(p.currentContainer.Children, newToken)
	} else {
		p.tree = append(p.tree, newToken)
	}

	p.currentContainer = newToken
	p.currentToken = newToken
	p.pushState(ParserStateJSONStart)
	p.depth++
}

func (p *JSONParser) startNewJSONField() {
	fieldName := strings.TrimSpace(p.buffer.String())
	fieldName = strings.Trim(fieldName, "\"'")
	newToken := &Token{Type: TokenTypeJSONField, Content: fieldName, Pos: p.tokenStart, Children: make([]*Token, 0)}
	p.currentContainer.Children = append(p.currentContainer.Children, newToken)
	p.currentToken = newToken
	p.buffer.Reset()
}

func (p *JSONParser) findParentContainer() *Token {
	if p.currentContainer == nil {
		return nil
	}

	var findParent func(*Token) *Token
	findParent = func(token *Token) *Token {
		for _, child := range token.Children {
			if child == p.currentContainer {
				return token
			}
			if result := findParent(child); result != nil {
				return result
			}
		}

		return nil
	}

	for _, token := range p.tree {
		if result := findParent(token); result != nil {
			return result
		}
	}

	return nil
}

func (p *JSONParser) autoCloseJSON() {
	for p.depth > 0 {
		p.depth--
		p.completeCurrentToken()
	}
}

func (p *JSONParser) flushBuffer() {
	if p.buffer.Len() > 0 {
		content := p.buffer.String()
		if !p.insideJSON {
			p.tree = append(p.tree, &Token{Type: TokenTypeText, Content: content, Pos: p.tokenStart})
		}

		p.buffer.Reset()
	}
}

func (p *JSONParser) getCompletedTokens() []*Token {
	var completedTokens []*Token

	for _, token := range p.tree {
		if token.Type == TokenTypeText || (token.Type == TokenTypeJSONObject || token.Type == TokenTypeJSONArray) && p.state == ParserStateText {
			completedTokens = append(completedTokens, token)
		} else {
			break
		}
	}

	p.tree = p.tree[len(completedTokens):]

	return completedTokens
}

func (p *JSONParser) updatePosition(char rune) {
	p.pos.Offset++
	if char == '\n' {
		p.pos.Line++
		p.pos.Column = 1
	} else {
		p.pos.Column++
	}
}
