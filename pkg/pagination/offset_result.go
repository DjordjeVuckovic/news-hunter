package pagination

// OffsetResult represents traditional offset-based pagination
// Kept for backward compatibility
type OffsetResult[T any] struct {
	Items   []T   `json:"items"`
	Total   int64 `json:"total"`
	Page    int   `json:"page"`
	Size    int   `json:"size"`
	HasMore bool  `json:"has_more"`
}

// NewOffsetResult creates a new offset-based result
func NewOffsetResult[T any](items []T, total int64, page int, size int) *OffsetResult[T] {
	offset := (page - 1) * size
	hasMore := int64(offset+size) < total

	return &OffsetResult[T]{
		Items:   items,
		Total:   total,
		Page:    page,
		Size:    size,
		HasMore: hasMore,
	}
}
