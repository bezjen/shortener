package repository

type Repository interface {
	Save(id string, url string)
	GetByShortURL(id string) string
}
