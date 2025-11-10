package pg

import (
	"fmt"

	"github.com/DjordjeVuckovic/news-hunter/internal/domain/operator"
	"github.com/DjordjeVuckovic/news-hunter/internal/domain/query"
)

// buildSearchVector returns the tsvector expression for full-text search
// Always uses the pre-computed 'search_vector' column which has weights:
//   - title: 'A' (weight 1.0)
//   - description: 'B' (weight 0.4)
//   - subtitle/content: 'C' (weight 0.2)
//   - author: 'D' (weight 0.1)
//
// The pre-computed column is GIN-indexed for fast searches.
func buildSearchVector(fields []string, weights map[string]float64, lang query.Language) string {
	// Always use pre-computed weighted search_vector
	// It's pre-indexed and has field weights baked in
	// Custom field weights are applied via ts_rank weights array instead
	return "search_vector"
}

// buildTsQuery constructs a PostgreSQL tsquery expression based on operator
// paramNum: The parameter number to use ($1, $2, etc.)
// Returns: "plainto_tsquery('english'::regconfig, $1)" or "websearch_to_tsquery(...)"
func buildTsQuery(op operator.Operator, lang query.Language, paramNum int) string {
	if lang == "" {
		lang = query.LanguageEnglish
	}

	if op.IsOr() {
		// websearch_to_tsquery treats space-separated terms as AND
		// "climate change" -> "climate & change"
		return fmt.Sprintf("websearch_to_tsquery('%s'::regconfig, $%d)", lang, paramNum)
	}

	// plainto_tsquery uses AND by default
	// "climate change" -> "climate | change"
	return fmt.Sprintf("plainto_tsquery('%s'::regconfig, $%d)", lang, paramNum)
}

// buildRankExpression constructs a ts_rank expression with custom field weights
// The pre-computed search_vector has weights: title=A, description=B, subtitle/content=C, author=D
// PostgreSQL's default weight values are: A=1.0, B=0.4, C=0.2, D=0.1
// If custom DefaultFieldWeights are specified, we override these defaults via the weights array
// Returns: "ts_rank('{3.0, 2.0, 1.0, 0.1}', search_vector, query)" or default "ts_rank(search_vector, query)"
func buildRankExpression(fields []string, weights map[string]float64, lang query.Language, op operator.Operator, paramNum int) string {
	vectorExpr := buildSearchVector(fields, weights, lang)
	queryExpr := buildTsQuery(op, lang, paramNum)

	// Map field names to PostgreSQL weight labels (A, B, C, D)
	//   title → A (weights[0])
	//   description → B (weights[1])
	//   subtitle/content → C (weights[2])
	//   author → D (weights[3])
	fieldToLabel := map[string]int{
		"title":       0, // A
		"description": 1, // B
		"subtitle":    2, // C
		"content":     2, // C
		"author":      3, // D
	}

	// Default PostgreSQL weights for A, B, C, D
	weightArray := []float64{1.0, 0.4, 0.2, 0.1}

	// Apply custom weights if specified
	if len(weights) > 0 {
		for field, weight := range weights {
			if labelIdx, ok := fieldToLabel[field]; ok {
				weightArray[labelIdx] = weight
			}
		}

		// Use custom weights array in ts_rank
		return fmt.Sprintf("ts_rank('{%.2f, %.2f, %.2f, %.2f}', %s, %s)",
			weightArray[0], weightArray[1], weightArray[2], weightArray[3],
			vectorExpr, queryExpr)
	}

	// Use default weights (let PostgreSQL use its defaults)
	return fmt.Sprintf("ts_rank(%s, %s)", vectorExpr, queryExpr)
}

// buildTsWhereClause constructs the WHERE clause for full-text search
// Always uses the pre-computed weighted search_vector with GIN index
func buildTsWhereClause(fields []string, weights map[string]float64, lang query.Language, op operator.Operator, paramNum int) string {
	vectorExpr := buildSearchVector(fields, weights, lang) // Returns "search_vector"
	queryExpr := buildTsQuery(op, lang, paramNum)

	return fmt.Sprintf("%s @@ %s", vectorExpr, queryExpr)
}
