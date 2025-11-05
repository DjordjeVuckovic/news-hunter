package dto

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Cursor represents a position in a search result set
// It contains the relevance score and ID of the last item
type Cursor struct {
	Score float64   `json:"s"` // Raw score for pagination consistency
	ID    uuid.UUID `json:"i"`
}

// EncodeCursor converts a Cursor to a base64-encoded string
func EncodeCursor(score float64, id uuid.UUID) (string, error) {
	if id == uuid.Nil {
		return "", fmt.Errorf("cursor ID cannot be nil")
	}

	c := Cursor{
		Score: score,
		ID:    id,
	}

	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// DecodeCursor parses a base64-encoded cursor string
func DecodeCursor(s string) (*Cursor, error) {
	if s == "" {
		return nil, nil
	}

	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cursor: %w", err)
	}

	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cursor: %w", err)
	}

	if c.ID == uuid.Nil {
		return nil, fmt.Errorf("invalid cursor: ID cannot be nil")
	}

	return &c, nil
}

// MustEncodeCursor is like EncodeCursor but panics on error
// Use only when you're certain the inputs are valid
func MustEncodeCursor(score float64, id uuid.UUID) string {
	cursor, err := EncodeCursor(score, id)
	if err != nil {
		panic(err)
	}
	return cursor
}
