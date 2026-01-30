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
