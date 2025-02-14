package store

type Store interface {
	Get(shortURL string) (string, bool)
	Set(shortURL, originalURL string)
}
