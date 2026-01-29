package parser

import "testing"

func TestBoolTokenizer_Tokenize(t *testing.T) {
	input := "A AND B OR (C NOT D)"
	tokenizer := NewBoolTokenizer(input)
	tokens := tokenizer.Tokenize(input)

	expectedTypes := []TokenType{
		WORD, AND, WORD, OR, LPAREN, WORD, NOT, WORD, RPAREN, EOF,
	}

	if len(tokens) != len(expectedTypes) {
		t.Fatalf("expected %d tokens, got %d", len(expectedTypes), len(tokens))
	}
}
