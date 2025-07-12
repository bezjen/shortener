package repository

type Repository interface {
	Save(id string, url string)
	GetByShortUrl(id string) string
}
