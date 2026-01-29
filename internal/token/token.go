package token

type Type int

const (
	EOF Type = iota
	WORD
	AND
	OR
	NOT
	LPAREN
	RPAREN
)

func (t Type) String() string {
	switch t {
	case EOF:
		return "EOF"
	case WORD:
		return "WORD"
	case AND:
		return "AND"
	case OR:
		return "OR"
	case NOT:
		return "NOT"
	case LPAREN:
		return "LPAREN"
	case RPAREN:
		return "RPAREN"
	default:
		return "UNKNOWN"
	}
}

// Token represents a lexical token with its type and literal value.
type Token struct {
	Type  Type
	Value string
}
