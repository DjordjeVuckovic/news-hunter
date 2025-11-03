package pagination

// CursorResult represents a cursor-based paginated result
// Generic type T allows reuse across different entity types
type CursorResult[T any] struct {
	Items      []T     `json:"items"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

// NewCursorResult creates a new cursor-based result
// If there are more items than requested (size+1), it:
// - Returns only the requested number of items
// - Sets HasMore to true
// - Generates NextCursor from the last returned item
func NewCursorResult[T any](items []T, size int, cursorFn func(T) (string, error)) (*CursorResult[T], error) {
	hasMore := len(items) > size

	// Trim to requested size if we fetched size+1
	if hasMore {
		items = items[:size]
	}

	result := &CursorResult[T]{
		Items:   items,
		HasMore: hasMore,
	}

	// Generate cursor from last item if there are more results
	if hasMore && len(items) > 0 {
		cursor, err := cursorFn(items[len(items)-1])
		if err != nil {
			return nil, err
		}
		result.NextCursor = &cursor
	}

	return result, nil
}
