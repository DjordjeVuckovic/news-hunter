package query

type Capabilities struct {
	Match      bool
	MultiMatch bool
	Phrase     bool
	Fuzzy      bool
}
