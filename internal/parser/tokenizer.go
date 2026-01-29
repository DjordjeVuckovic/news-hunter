package parser

// Tokenizer interface defines the method for tokenizing input strings.
type Tokenizer interface {
	Tokenize(input string) []Token
}
