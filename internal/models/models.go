// Package models contains data structures used throughout the URL shortener service.
// It defines request and response types for API endpoints, as well as internal representations
// of URLs and batch operations.
package models

// ShortenRequest represents a request to shorten a URL.
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse represents a response containing the shortened URL.
type ShortenResponse struct {
	Result string `json:"result"`
}

// BatchRequest is a slice of BatchRequestItem for batch shortening requests.
type BatchRequest []BatchRequestItem

// BatchRequestItem represents a single item in a batch shorten request.
type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponse is a slice of BatchResponseItem for batch shortening responses.
type BatchResponse []BatchResponseItem

// BatchResponseItem represents a single item in a batch shorten response.
type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// URL represents a shortened URL mapping with metadata.
type URL struct {
	UUID        string `db:"id" json:"-"`
	UserID      string `db:"user_id" json:"-"`
	ShortURL    string `db:"short_url" json:"short_url"`
	OriginalURL string `db:"original_url" json:"original_url"`
	IsDeleted   bool   `db:"is_deleted" json:"-"`
}
