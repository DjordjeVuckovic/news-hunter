package parser

import (
	"testing"
)

func TestBoolTokenizer_Tokenize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "empty input",
			input: "",
			expected: []Token{
				{Type: EOF},
			},
		},
		{
			name:  "single word",
			input: "climate",
			expected: []Token{
				{Type: WORD, Value: "climate"},
				{Type: EOF},
			},
		},
		{
			name:  "simple AND expression",
			input: "climate AND change",
			expected: []Token{
				{Type: WORD, Value: "climate"},
				{Type: AND, Value: "AND"},
				{Type: WORD, Value: "change"},
				{Type: EOF},
			},
		},
		{
			name:  "simple OR expression",
			input: "renewable OR sustainable",
			expected: []Token{
				{Type: WORD, Value: "renewable"},
				{Type: OR, Value: "OR"},
				{Type: WORD, Value: "sustainable"},
				{Type: EOF},
			},
		},
		{
			name:  "NOT expression",
			input: "climate NOT politics",
			expected: []Token{
				{Type: WORD, Value: "climate"},
				{Type: NOT, Value: "NOT"},
				{Type: WORD, Value: "politics"},
				{Type: EOF},
			},
		},
		{
			name:  "parenthesized expression",
			input: "(renewable OR sustainable) AND energy",
			expected: []Token{
				{Type: LPAREN, Value: "("},
				{Type: WORD, Value: "renewable"},
				{Type: OR, Value: "OR"},
				{Type: WORD, Value: "sustainable"},
				{Type: RPAREN, Value: ")"},
				{Type: AND, Value: "AND"},
				{Type: WORD, Value: "energy"},
				{Type: EOF},
			},
		},
		{
			name:  "complex expression from plan",
			input: "A AND B OR (C NOT D)",
			expected: []Token{
				{Type: WORD, Value: "A"},
				{Type: AND, Value: "AND"},
				{Type: WORD, Value: "B"},
				{Type: OR, Value: "OR"},
				{Type: LPAREN, Value: "("},
				{Type: WORD, Value: "C"},
				{Type: NOT, Value: "NOT"},
				{Type: WORD, Value: "D"},
				{Type: RPAREN, Value: ")"},
				{Type: EOF},
			},
		},
		{
			name:  "case insensitive operators",
			input: "climate and change or warming not politics",
			expected: []Token{
				{Type: WORD, Value: "climate"},
				{Type: AND, Value: "and"},
				{Type: WORD, Value: "change"},
				{Type: OR, Value: "or"},
				{Type: WORD, Value: "warming"},
				{Type: NOT, Value: "not"},
				{Type: WORD, Value: "politics"},
				{Type: EOF},
			},
		},
		{
			name:  "word starting with operator prefix is not split",
			input: "ANDROID OR ORNAMENT OR NOTHING",
			expected: []Token{
				{Type: WORD, Value: "ANDROID"},
				{Type: OR, Value: "OR"},
				{Type: WORD, Value: "ORNAMENT"},
				{Type: OR, Value: "OR"},
				{Type: WORD, Value: "NOTHING"},
				{Type: EOF},
			},
		},
		{
			name:  "quoted phrase",
			input: `"climate change" AND energy`,
			expected: []Token{
				{Type: WORD, Value: "climate change"},
				{Type: AND, Value: "AND"},
				{Type: WORD, Value: "energy"},
				{Type: EOF},
			},
		},
		{
			name:  "unclosed quote reads to end",
			input: `"climate change`,
			expected: []Token{
				{Type: WORD, Value: "climate change"},
				{Type: EOF},
			},
		},
		{
			name:  "nested parentheses",
			input: "((A OR B) AND C)",
			expected: []Token{
				{Type: LPAREN, Value: "("},
				{Type: LPAREN, Value: "("},
				{Type: WORD, Value: "A"},
				{Type: OR, Value: "OR"},
				{Type: WORD, Value: "B"},
				{Type: RPAREN, Value: ")"},
				{Type: AND, Value: "AND"},
				{Type: WORD, Value: "C"},
				{Type: RPAREN, Value: ")"},
				{Type: EOF},
			},
		},
		{
			name:  "whitespace only",
			input: "   \t\n  ",
			expected: []Token{
				{Type: EOF},
			},
		},
		{
			name:  "extra whitespace between tokens",
			input: "climate   AND   change",
			expected: []Token{
				{Type: WORD, Value: "climate"},
				{Type: AND, Value: "AND"},
				{Type: WORD, Value: "change"},
				{Type: EOF},
			},
		},
		{
			name:  "words with underscores and digits",
			input: "field_1 AND test_value2",
			expected: []Token{
				{Type: WORD, Value: "field_1"},
				{Type: AND, Value: "AND"},
				{Type: WORD, Value: "test_value2"},
				{Type: EOF},
			},
		},
	}

	tokenizer := NewBoolTokenizer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenizer.Tokenize(tt.input)

			if len(tokens) != len(tt.expected) {
				t.Fatalf("expected %d tokens, got %d\nexpected: %v\ngot:      %v", len(tt.expected), len(tokens), tt.expected, tokens)
			}

			for i, tok := range tokens {
				if tok.Type != tt.expected[i].Type {
					t.Errorf("token[%d] type: expected %s, got %s", i, tt.expected[i].Type, tok.Type)
				}
				if tok.Value != tt.expected[i].Value {
					t.Errorf("token[%d] value: expected %q, got %q", i, tt.expected[i].Value, tok.Value)
				}
			}
		})
	}
}

func TestBoolTokenizer_IsReusable(t *testing.T) {
	tokenizer := NewBoolTokenizer()

	first := tokenizer.Tokenize("A AND B")
	second := tokenizer.Tokenize("C OR D")

	if len(first) != 4 {
		t.Fatalf("first call: expected 4 tokens, got %d", len(first))
	}
	if len(second) != 4 {
		t.Fatalf("second call: expected 4 tokens, got %d", len(second))
	}
	if second[0].Value != "C" {
		t.Errorf("second call first token: expected %q, got %q", "C", second[0].Value)
	}
}
