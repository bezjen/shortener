package repository

type InMemoryRepository struct {
	storage map[string]string
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(map[string]string),
	}
}

func (m *InMemoryRepository) Save(shortURL string, url string) {
	m.storage[shortURL] = url
}

func (m *InMemoryRepository) GetByShortURL(shortURL string) string {
	return m.storage[shortURL]
}
