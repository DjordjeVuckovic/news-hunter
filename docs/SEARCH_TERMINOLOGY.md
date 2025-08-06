# Search Types

1. Simple/Basic Search

- "Keyword Search" or "Full-text Search"
- Single terms or phrases
- Example: "climate change"

2. Boolean Search

- "Boolean Query" or "Boolean Search"
- Uses logical operators: AND, OR, NOT
- Example: "climate AND change OR global warming"

3. Advanced Query Language

- "Query DSL" (Domain Specific Language)
- "Structured Query"
- "Query Parser" (the component that interprets it)

Common Industry Terms

Query Types:

- Simple Query: Basic keyword matching
- Boolean Query: Logical operators (AND, OR, NOT)
- Phrase Query: Exact phrase matching ("exact phrase")
- Wildcard Query: Pattern matching (clim*, ?ange)
- Fuzzy Query: Approximate matching (typo tolerance)
- Range Query: Date/numeric ranges

Search Features:

- Full-Text Search (FTS): Search through document content
- Faceted Search: Filter by categories/attributes
- Autocomplete/Typeahead: Search suggestions
- Semantic Search: Meaning-based search
- Relevance Scoring: Ranking results by relevance

For Your API Documentation:

// SearchRequest represents different search types
type SearchRequest struct {
Query string `json:"query"`
Type  string `json:"type"` // "simple", "boolean", "advanced"
}

// Examples:
// Simple: "climate change"
// Boolean: "climate AND change OR warming"
// Advanced: "title:climate AND content:change"

Standard Elasticsearch/Lucene Terminology:

- Match Query: Simple text search
- Bool Query: Boolean logic with must/should/must_not
- Query String Query: User-friendly query parser
- Multi-Match Query: Search across multiple fields

For your API, I'd recommend calling them:
- "Simple Search" (basic keywords)
- "Boolean Search" (AND/OR operators)
- "Advanced Search" (field-specific, complex queries)
