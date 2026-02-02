package suite

import (
	"fmt"
	"os"

	"github.com/DjordjeVuckovic/news-hunter/internal/types/operator"
	"github.com/DjordjeVuckovic/news-hunter/internal/types/query"
	"gopkg.in/yaml.v3"
)

func LoadFromFile(path string) (*TestSuite, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read suite file: %w", err)
	}
	return Parse(data)
}

func Parse(data []byte) (*TestSuite, error) {
	var s TestSuite
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse suite YAML: %w", err)
	}
	if len(s.Queries) == 0 {
		return nil, fmt.Errorf("suite has no queries")
	}
	for i, q := range s.Queries {
		if q.ID == "" {
			return nil, fmt.Errorf("query at index %d has no id", i)
		}
		if q.Kind == "" {
			return nil, fmt.Errorf("query %q has no kind", q.ID)
		}
	}
	return &s, nil
}

func ToDomainQuery(bq *BenchmarkQuery) (*query.Base, error) {
	base := &query.Base{Kind: bq.Kind}

	switch bq.Kind {
	case query.QueryStringType:
		if bq.QueryString == nil {
			return nil, fmt.Errorf("query %q: kind is %s but query_string spec is nil", bq.ID, bq.Kind)
		}
		opts := []query.StringOption{}
		if bq.QueryString.Language != "" {
			opts = append(opts, query.WithQueryStringLanguage(query.Language(bq.QueryString.Language)))
		}
		if bq.QueryString.Operator != "" {
			op, err := operator.Parse(bq.QueryString.Operator)
			if err != nil {
				return nil, fmt.Errorf("query %q: invalid operator: %w", bq.ID, err)
			}
			opts = append(opts, query.WithQueryStringOperator(op))
		}
		base.QueryString = query.NewQueryString(bq.QueryString.Query, opts...)

	case query.MatchType:
		if bq.Match == nil {
			return nil, fmt.Errorf("query %q: kind is %s but match spec is nil", bq.ID, bq.Kind)
		}
		opts := []query.MatchQueryOption{}
		if bq.Match.Language != "" {
			opts = append(opts, query.WithMatchLanguage(query.Language(bq.Match.Language)))
		}
		if bq.Match.Operator != "" {
			op, err := operator.Parse(bq.Match.Operator)
			if err != nil {
				return nil, fmt.Errorf("query %q: invalid operator: %w", bq.ID, err)
			}
			opts = append(opts, query.WithMatchOperator(op))
		}
		if bq.Match.Fuzziness != "" {
			opts = append(opts, query.WithMatchFuzziness(bq.Match.Fuzziness))
		}
		base.Match = query.NewMatch(bq.Match.Field, bq.Match.Query, opts...)

	case query.MultiMatchType:
		if bq.MultiMatch == nil {
			return nil, fmt.Errorf("query %q: kind is %s but multi_match spec is nil", bq.ID, bq.Kind)
		}
		opts := []query.MultiMatchQueryOption{}
		if bq.MultiMatch.Language != "" {
			opts = append(opts, query.WithMultiMatchLanguage(query.Language(bq.MultiMatch.Language)))
		}
		if bq.MultiMatch.Operator != "" {
			op, err := operator.Parse(bq.MultiMatch.Operator)
			if err != nil {
				return nil, fmt.Errorf("query %q: invalid operator: %w", bq.ID, err)
			}
			opts = append(opts, query.WithMultiMatchOperator(op))
		}
		mm, err := query.NewMultiMatchQuery(bq.MultiMatch.Query, bq.MultiMatch.Fields, opts...)
		if err != nil {
			return nil, fmt.Errorf("query %q: %w", bq.ID, err)
		}
		base.MultiMatch = mm

	case query.PhraseType:
		if bq.Phrase == nil {
			return nil, fmt.Errorf("query %q: kind is %s but phrase spec is nil", bq.ID, bq.Kind)
		}
		opts := []query.PhraseOption{}
		if bq.Phrase.Language != "" {
			opts = append(opts, query.WithPhraseLanguage(query.Language(bq.Phrase.Language)))
		}
		if bq.Phrase.Slop > 0 {
			opts = append(opts, query.WithPhraseSlop(bq.Phrase.Slop))
		}
		ph, err := query.NewPhrase(bq.Phrase.Query, bq.Phrase.Fields, opts...)
		if err != nil {
			return nil, fmt.Errorf("query %q: %w", bq.ID, err)
		}
		base.Phrase = ph

	case query.BooleanType:
		if bq.Boolean == nil {
			return nil, fmt.Errorf("query %q: kind is %s but boolean spec is nil", bq.ID, bq.Kind)
		}
		b := &query.Boolean{
			Expression: bq.Boolean.Expression,
		}
		if bq.Boolean.Language != "" {
			b.Language = query.Language(bq.Boolean.Language)
		}
		base.Boolean = b

	default:
		return nil, fmt.Errorf("query %q: unsupported kind %q", bq.ID, bq.Kind)
	}

	return base, nil
}
