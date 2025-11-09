package dto

type SearchQuery interface {
	isQuery()
}

type SearchRequest struct {
	Query SearchQuery `json:"query"`
	Size  int         `json:"size"`
}

type MatchQueryWrapper struct {
	Match map[string]MatchQueryParams `json:"match"`
}

func (MatchQueryWrapper) isQuery() {}

type MultiMatchQueryWrapper struct {
	MultiMatch MultiMatchQueryParams `json:"multi_match"`
}

func (MultiMatchQueryWrapper) isQuery() {}

type BoolQueryWrapper struct {
	Bool BoolQueryParams `json:"bool"`
}

func (BoolQueryWrapper) isQuery() {}

type MatchQueryParams struct {
	Query     string `json:"query" validate:"required,min=1"`
	Operator  string `json:"operator,omitempty"`
	Fuzziness string `json:"fuzziness,omitempty"`
	// Note: Field is in the map key, not here
}

type MultiMatchQueryParams struct {
	Query        string             `json:"query" validate:"required,min=1"`
	Fields       []string           `json:"fields" validate:"required,min=1"`
	FieldWeights map[string]float64 `json:"field_weights,omitempty"`
	Operator     string             `json:"operator,omitempty"`
}

type BoolQueryParams struct {
	Must    []any `json:"must,omitempty"`
	Should  []any `json:"should,omitempty"`
	MustNot []any `json:"must_not,omitempty"`
	Filter  []any `json:"filter,omitempty"`
}
