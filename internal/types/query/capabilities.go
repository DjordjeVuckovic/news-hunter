package query

// Capabilities reports which search paradigms the running backend exposes.
type Capabilities struct {
	StringQuery bool `json:"string_query"`
	Match       bool `json:"match"`
	MultiMatch  bool `json:"multi_match"`
	Phrase      bool `json:"phrase"`
	Boolean     bool `json:"boolean"`
	Semantic    bool `json:"semantic"`
}
