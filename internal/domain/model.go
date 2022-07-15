package domain

type APIOriginBatch struct {
	ID     string `json:"correlation_id"`
	Origin string `json:"original_url"`
}

type APIShortBatch struct {
	ID       string `json:"correlation_id"`
	ShortURL string `json:"short_url"`
}

type URLForAddStorage struct {
	IDUser   string
	ShortURL string
	FullURL  string
	Deleted  bool
}

type URLForGetAll struct {
	ShortURL string
	FullURL  string
}

type URLpair struct {
	ShortURL string `json:"short_url"`
	FullURL  string `json:"original_url"`
}
