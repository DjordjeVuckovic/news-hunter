package es

import "time"

// Document ESDocument represents the document structure for Elasticsearch
type Document struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Subtitle    string    `json:"subtitle"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	URL         string    `json:"url"`
	Language    string    `json:"language"`
	CreatedAt   time.Time `json:"created_at"`
	SourceId    string    `json:"source_id"`
	SourceName  string    `json:"source_name"`
	PublishedAt time.Time `json:"published_at"`
	Category    string    `json:"category"`
	ImportedAt  time.Time `json:"imported_at"`
	IndexedAt   time.Time `json:"indexed_at"`
}
