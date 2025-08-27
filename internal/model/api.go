package model

type ShortenJSONRequest struct {
	URL string `json:"url"`
}

type ShortenJSONResponse struct {
	ShortURL string `json:"result,omitempty"`
	Error    string `json:"error,omitempty"`
}

type ShortenBatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenBatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type UserURLResponseItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewShortenBatchRequestItem(correlationID string, OriginalURL string) *ShortenBatchRequestItem {
	return &ShortenBatchRequestItem{
		CorrelationID: correlationID,
		OriginalURL:   OriginalURL,
	}
}

func NewShortenBatchResponseItem(correlationID string, shortURL string) *ShortenBatchResponseItem {
	return &ShortenBatchResponseItem{
		CorrelationID: correlationID,
		ShortURL:      shortURL,
	}
}

func NewUserURLResponseItem(shortURL string, originalURL string) *UserURLResponseItem {
	return &UserURLResponseItem{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
}
