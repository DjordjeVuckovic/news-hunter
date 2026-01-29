package pg

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/DjordjeVuckovic/news-hunter/internal/parser"
)

type BooleanParser struct {
	tokenizer *parser.BoolTokenizer
}

func NewBooleanParser() *BooleanParser {
	return &BooleanParser{
		tokenizer: parser.NewBoolTokenizer(),
	}
}

func (p *BooleanParser) Parse(expression string) (string, error) {
	tokens := p.tokenizer.Tokenize(expression)
	return convertToTsquery(tokens)
}

func convertToTsquery(tokens []parser.Token) (string, error) {
	var parts []string
	prevType := parser.EOF

	for _, tok := range tokens {
		if tok.Type == parser.EOF {
			break
		}

		if needsImplicitAnd(prevType, tok.Type) {
			parts = append(parts, "&")
		}

		switch tok.Type {
		case parser.WORD:
			words := strings.Fields(sanitizeTerm(tok.Value))
			if len(words) > 1 {
				parts = append(parts, strings.Join(words, " <-> "))
			} else if len(words) == 1 {
				parts = append(parts, words[0])
			}
		case parser.AND:
			parts = append(parts, "&")
		case parser.OR:
			parts = append(parts, "|")
		case parser.NOT:
			parts = append(parts, "!")
		case parser.LPAREN:
			parts = append(parts, "(")
		case parser.RPAREN:
			parts = append(parts, ")")
		}

		prevType = tok.Type
	}

	result := strings.Join(parts, " ")
	if result == "" {
		return "", fmt.Errorf("empty boolean expression")
	}
	return result, nil
}

func needsImplicitAnd(prev, curr parser.TokenType) bool {
	prevIsValue := prev == parser.WORD || prev == parser.RPAREN
	currIsValue := curr == parser.WORD || curr == parser.LPAREN || curr == parser.NOT
	return prevIsValue && currIsValue
}

func sanitizeTerm(word string) string {
	var b strings.Builder
	for _, r := range word {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == ' ' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
