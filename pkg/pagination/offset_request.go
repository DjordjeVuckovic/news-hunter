package pagination

// OffsetRequest represents an offset-based pagination request
type OffsetRequest struct {
	Page int `json:"page" query:"page" validate:"min=1"`
	Size int `json:"size" query:"size" validate:"min=1,max=100"`
}

// Validate validates and normalizes offset pagination parameters
func (r *OffsetRequest) Validate() error {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Size <= 0 {
		r.Size = PageDefaultSize
	}
	if r.Size > PageMaxSize {
		r.Size = PageMaxSize
	}
	return nil
}
