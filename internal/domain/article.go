package domain

import (
	"github.com/google/uuid"
	"net/url"
	"time"
)

type Article struct {
	ID          uuid.UUID       `json:"id"`
	Title       string          `json:"title"`
	Subtitle    string          `json:"subtitle,omitempty"`
	Content     string          `json:"content"`
	Author      string          `json:"author,omitempty"`
	Description string          `json:"description,omitempty"`
	Language    string          `json:"language,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	URL         url.URL         `json:"sourceUrl,omitempty"`
	Metadata    ArticleMetadata `json:"metadata"`
}

type ArticleMetadata struct {
	SourceId    string    `json:"sourceId"`
	SourceName  string    `json:"sourceName"`
	PublishedAt time.Time `json:"publishedAt"`
}
