package parser

import (
	"strings"
	"unicode"
)

type BoolTokenizer struct {
	input []rune
	pos   int
}

func NewBoolTokenizer() *BoolTokenizer {
	return &BoolTokenizer{}
}

func (t *BoolTokenizer) Tokenize(input string) []Token {
	t.input = []rune(input)
	t.pos = 0

	var tokens []Token

	for t.pos < len(t.input) {
		ch := t.input[t.pos]
		switch {
		case ch == '(':
			tokens = append(tokens, Token{Type: LPAREN, Value: "("})
			t.pos++
		case ch == ')':
			tokens = append(tokens, Token{Type: RPAREN, Value: ")"})
			t.pos++
		case ch == '"':
			tokens = append(tokens, t.readQuoted())
		case isWordChar(ch):
			tokens = append(tokens, t.readWord())
		default:
			t.pos++
		}
		t.skipWhitespace()
	}

	tokens = append(tokens, Token{Type: EOF})
	return tokens
}

func (t *BoolTokenizer) skipWhitespace() {
	for t.pos < len(t.input) && unicode.IsSpace(t.input[t.pos]) {
		t.pos++
	}
}

func (t *BoolTokenizer) readWord() Token {
	start := t.pos
	for t.pos < len(t.input) && isWordChar(t.input[t.pos]) {
		t.pos++
	}

	word := string(t.input[start:t.pos])

	switch strings.ToUpper(word) {
	case "AND":
		return Token{Type: AND, Value: word}
	case "OR":
		return Token{Type: OR, Value: word}
	case "NOT":
		return Token{Type: NOT, Value: word}
	default:
		return Token{Type: WORD, Value: word}
	}
}

func (t *BoolTokenizer) readQuoted() Token {
	t.pos++ // skip opening quote
	start := t.pos
	for t.pos < len(t.input) && t.input[t.pos] != '"' {
		t.pos++
	}
	value := string(t.input[start:t.pos])
	if t.pos < len(t.input) {
		t.pos++ // skip closing quote
	}
	return Token{Type: WORD, Value: value}
}

func isWordChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}
