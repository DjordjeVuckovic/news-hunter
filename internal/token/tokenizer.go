package token

// Tokenizer interface defines the method for tokenizing input strings.
type Tokenizer interface {
	Tokenize(input string) []Token
}

type Validator interface {
	Validate(tokens []Token) error
}
