package pagination

// CursorResult represents a cursor-based paginated result
// Generic type T allows reuse across different entity types
type CursorResult[T any] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}
