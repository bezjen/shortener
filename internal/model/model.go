package model

type PostShortURLJSONRequest struct {
	URL string `json:"url"`
}

type PostShortURLJSONResponse struct {
	ShortURL string `json:"result"`
}
