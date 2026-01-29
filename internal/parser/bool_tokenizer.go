package parser

import "unicode"

// BoolTokenizer is responsible for breaking input strings into tokens for parsing. Mainly used for
type BoolTokenizer struct {
	input string
	pos   int
}

// NewBoolTokenizer creates a new Tokenizer for the given input string.
func NewBoolTokenizer(input string) *BoolTokenizer {
	return &BoolTokenizer{input: input, pos: 0}
}

// Tokenize processes the input string and returns a slice of Tokens.
func (t *BoolTokenizer) Tokenize() []Token {
	if len(t.input) == 0 {
		return []Token{{Type: EOF, Value: ""}}
	}

	tokens := make([]Token, 0)

	for t.pos < len(t.input) {
		t.skipWhitespace()
		char := t.input[t.pos]
		switch char {
		case '(':
			t.pos++
			tokens = append(tokens, Token{Type: LPAREN, Value: "("})
		case ')':
			t.pos++
			tokens = append(tokens, Token{Type: RPAREN, Value: ")"})
		case 'A':
			if len(t.input[t.pos:]) >= 3 && t.input[t.pos:t.pos+3] == "AND" {
				t.pos += 3
				tokens = append(tokens, Token{Type: AND, Value: t.input[t.pos : t.pos+3]})
				continue
			}
			wordToken := t.readWord()
			if wordToken != nil {
				tokens = append(tokens, *wordToken)
			} else {
				t.pos++
			}
		case 'N':
			if len(t.input[t.pos:]) >= 3 && t.input[t.pos:t.pos+3] == "NOT" {
				t.pos += 3
				tokens = append(tokens, Token{Type: NOT, Value: "NOT"})
				continue
			}
			wordToken := t.readWord()
			if wordToken != nil {
				tokens = append(tokens, *wordToken)
			} else {
				t.pos++
			}
		case 'O':
			if len(t.input[t.pos:]) >= 2 && t.input[t.pos:t.pos+2] == "OR" {
				t.pos += 2
				tokens = append(tokens, Token{Type: OR, Value: "OR"})
				continue
			}
			wordToken := t.readWord()
			if wordToken != nil {
				tokens = append(tokens, *wordToken)
			} else {
				t.pos++
			}
		default:
			wordToken := t.readWord()
			if wordToken != nil {
				tokens = append(tokens, *wordToken)
			} else {
				t.pos++
			}
		}

	}

	tokens = append(tokens, Token{Type: EOF, Value: ""})

	return tokens
}

func (t *BoolTokenizer) skipWhitespace() {
	for t.pos < len(t.input) && (t.input[t.pos] == ' ' || t.input[t.pos] == '\t' || t.input[t.pos] == '\n' || t.input[t.pos] == '\r') {
		t.pos++
	}
}

func (t *BoolTokenizer) readWord() *Token {
	start := t.pos
	for t.pos < len(t.input) && t.isWordChar(t.input[t.pos]) {
		t.pos++
	}
	if start == t.pos {
		return nil
	}

	return &Token{Type: WORD, Value: t.input[start:t.pos]}
}

func (t *BoolTokenizer) isWordChar(char byte) bool {
	return unicode.IsLetter(rune(char)) || unicode.IsDigit(rune(char)) || char == '_' || char == '"'
}
