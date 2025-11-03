package pagination

// CursorRequest represents a cursor-based pagination request
type CursorRequest struct {
	Cursor *string `json:"cursor,omitempty" query:"cursor"`
	Size   int     `json:"size" query:"size" validate:"min=1,max=100"`
}

// Validate validates and normalizes cursor pagination parameters
func (r *CursorRequest) Validate() error {
	if r.Size <= 0 {
		r.Size = PageDefaultSize
	}
	if r.Size > PageMaxSize {
		r.Size = PageMaxSize
	}
	return nil
}
