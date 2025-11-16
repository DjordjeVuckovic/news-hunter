package query

type Capabilities struct {
	SupportsMatch      bool
	SupportsMultiMatch bool
	SupportsPhrase     bool
	SupportsFuzzy      bool
}
