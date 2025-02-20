package models

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type BatchRequest []BatchRequestItem

type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse []BatchResponseItem

type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type URL struct {
	UUID        string `db:"id" json:"-"`
	UserID      string `db:"user_id" json:"-"`
	ShortURL    string `db:"short_url" json:"short_url"`
	OriginalURL string `db:"original_url" json:"original_url"`
	IsDeleted   bool   `db:"is_deleted" json:"-"`
}
