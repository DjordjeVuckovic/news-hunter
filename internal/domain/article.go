package domain

import (
	"github.com/google/uuid"
	"net/url"
	"time"
)

const ArticleDefaultLanguage = "english"

type Article struct {
	ID          uuid.UUID       `json:"id"`
	Title       string          `json:"title"`
	Subtitle    string          `json:"subtitle,omitempty"`
	Content     string          `json:"content"`
	Author      string          `json:"author,omitempty"`
	Description string          `json:"description,omitempty"`
	Language    string          `json:"language,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	URL         url.URL         `json:"url,omitempty"`
	Metadata    ArticleMetadata `json:"metadata"`
}

type ArticleMetadata struct {
	// Essential source tracking
	SourceId    string    `json:"sourceId,omitempty"`
	SourceName  string    `json:"sourceName,omitempty"`
	PublishedAt time.Time `json:"publishedAt,omitempty"`
	// Content metadata
	Category string `json:"category,omitempty"`

	// System metadata
	ImportedAt time.Time `json:"importedAt,omitempty"`
}

type ArticleSearchResult struct {
	Article `json:"article" ` // Embedded Article struct for search results
	Rank    float32           `json:"rank"` // Rank based on search relevance
}
