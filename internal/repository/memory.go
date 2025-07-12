package repository

type InMemoryRepository struct {
	storage map[string]string
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(map[string]string),
	}
}

func (m *InMemoryRepository) Save(shortUrl string, url string) {
	m.storage[shortUrl] = url
}

func (m *InMemoryRepository) GetByShortUrl(shortUrl string) string {
	return m.storage[shortUrl]
}
