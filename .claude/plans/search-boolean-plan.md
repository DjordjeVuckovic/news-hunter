# SearchBoolean Implementation Plan

## Summary

Implement boolean search (`SearchBoolean`) for both PostgreSQL and Elasticsearch backends, supporting expressions like:
- `"climate AND change"`
- `"(renewable OR sustainable) AND energy"`
- `"Trump AND NOT biden"`

## Files to Modify

| File | Changes |
|------|---------|
| `internal/types/query/query.go` | Add `Language` field and helpers to `Boolean` struct |
| `internal/dto/query.go` | Add `BooleanParams` DTO, update `QueryWrapper` |
| `internal/router/search.go` | Add `handleBooleanQuery` handler, update switch |
| `internal/storage/pg/boolean_parser.go` | **NEW** - Boolean expression parser |
| `internal/storage/pg/searcher.go` | Implement `SearchBoolean` |
| `internal/storage/es/searcher.go` | Implement `SearchBoolean` |

## Parser Design: Tokenization vs AST

### Option 1: Simple String Replacement (NOT recommended)
```go
// Just replace operators
expr = strings.ReplaceAll(expr, " AND ", " & ")
expr = strings.ReplaceAll(expr, " OR ", " | ")
```
**Problems**: Doesn't handle quoted strings, case sensitivity, edge cases like "ANDY"

### Option 2: Tokenizer + Sequential Conversion (Recommended)
```go
// Tokenize into: [WORD:"climate", AND, LPAREN, WORD:"change", OR, WORD:"warming", RPAREN]
// Convert each token to tsquery syntax
```
**Pros**: Handles quoted strings, operators, parentheses correctly. No AST needed.
**Why it works**: PostgreSQL's `to_tsquery()` handles operator precedence internally.

### Option 3: Full AST Parser (Overkill)
```go
// Parse into tree structure, then generate tsquery from tree
type Node struct {
    Type     NodeType
    Value    string
    Left     *Node
    Right    *Node
}
```
**Problems**: Unnecessary complexity. We're not evaluating the expression, just translating syntax.

### Recommendation: Option 2 (Tokenizer)

We don't need a full AST because:
1. **PostgreSQL handles precedence** - `to_tsquery()` evaluates `&`, `|`, `!` with correct precedence
2. **We preserve parentheses** - User's grouping is maintained in output
3. **Simple translation** - Token-by-token conversion is sufficient

**Tokenizer output example**:
```
Input:  "climate AND (change OR warming)"
Tokens: [WORD:climate, AND, LPAREN, WORD:change, OR, WORD:warming, RPAREN]
Output: 'climate' & ('change' | 'warming')
```

## Implementation Steps

### Step 1: Extend `query.Boolean` Type

Add `Language` field and helper methods:

```go
type Boolean struct {
    Expression string   `json:"expression" validate:"required,min=1"`
    Language   Language `json:"language,omitempty"`
}

func NewBoolean(expression string, opts ...BooleanOption) (*Boolean, error)
func (q *Boolean) GetLanguage() Language
```

### Step 2: Add DTO Layer

**BooleanParams**:
```go
type BooleanParams struct {
    Expression string `json:"expression"`
    Language   string `json:"language,omitempty"`
}

func (p *BooleanParams) ToDomain() (*query.Boolean, error)
```

**Update QueryWrapper** - add `Boolean *BooleanParams` field

### Step 3: Add Router Handler

```go
case dquery.BooleanType:
    return r.handleBooleanQuery(c, req.Query.Boolean, opts)
```

### Step 4: Create Tokenizer-Based Parser

**New file**: `internal/storage/pg/boolean_parser.go`

```go
type tokenType int

const (
    tokenWord tokenType = iota
    tokenAnd
    tokenOr
    tokenNot
    tokenLeftParen
    tokenRightParen
)

type BooleanParser struct {
    input  string
    tokens []token
}

func (p *BooleanParser) Parse() (string, error) {
    p.tokenize()        // Break into tokens
    return p.convert()  // Convert tokens to tsquery syntax
}
```

**Tokenizer handles**:
- Quoted strings: `"climate change"` → single token
- Operators: AND, OR, NOT (case-insensitive)
- Symbols: &, |, !, (, )
- Terms: alphanumeric words

### Step 5: PostgreSQL Implementation

```go
func (r *Searcher) SearchBoolean(ctx context.Context, query *dquery.Boolean, baseOpts *dquery.BaseOptions) (*storage.SearchResult, error) {
    parser := NewBooleanParser(query.Expression)
    tsqueryStr, _ := parser.Parse()

    // Use to_tsquery with parsed string
    queryExpr := fmt.Sprintf("to_tsquery('%s'::regconfig, '%s')", lang, tsqueryStr)
}
```

### Step 6: Elasticsearch Implementation

Use `query_string` query (natively supports boolean syntax):

```go
queryStringQuery := &types.QueryStringQuery{
    Query:  query.Expression,  // Pass directly - ES parses it
    Fields: []string{"title^3.0", "description^2.0", "content^1.0"},
}
```

## API Example

```txt
POST /v1/articles/_search
{
  "size": 10,
  "query": {
    "boolean": {
      "expression": "(climate OR weather) AND change AND NOT politics",
      "language": "english"
    }
  }
}
```

## Verification

1. **Build**: `go build ./...`
2. **Unit tests**: `go test ./internal/storage/pg/... -run TestBoolean`
3. **Integration test**:
```bash
curl -X POST http://localhost:8080/v1/articles/_search \
  -H "Content-Type: application/json" \
  -d '{"size": 10, "query": {"boolean": {"expression": "climate AND change"}}}'
```
