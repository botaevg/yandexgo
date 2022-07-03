package domain

type APIOriginBatch struct {
	ID     string `json:"correlation_id"`
	Origin string `json:"original_url"`
}

type APIShortBatch struct {
	ID       string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}

type URL struct {
	IdUser   string
	ShortURL string
	FullURL  string
}
