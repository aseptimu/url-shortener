package service

import (
	"errors"
	"net/url"

	"github.com/aseptimu/url-shortener/internal/app/utils"
)

type Store interface {
	Get(shortURL string) (string, bool)
	Set(shortURL, originalURL string) (string, error)
	BatchSet(urls map[string]string) (map[string]string, error)
}

type URLShortener interface {
	ShortenURL(input string) (string, error)
	ShortenURLs(inputs []string) (map[string]string, error)
	GetOriginalURL(input string) (string, bool)
}

type URLService struct {
	store Store
}

func NewURLService(store Store) *URLService {
	return &URLService{store: store}
}

func (s *URLService) isValidURL(input string) bool {
	parsedURI, err := url.ParseRequestURI(input)
	return err == nil && parsedURI.Scheme != "" && parsedURI.Host != ""
}

var ErrConflict = errors.New("URL already exists")

func (s *URLService) ShortenURL(input string) (string, error) {
	if !s.isValidURL(input) {
		return "", errors.New("invalid URL format")
	}

	shortURL := utils.RandomString(6)
	storeURL, err := s.store.Set(shortURL, input)
	if err != nil {
		return "", err
	}

	if storeURL != shortURL {
		return storeURL, ErrConflict
	}

	return shortURL, nil
}

func (s *URLService) ShortenURLs(inputs []string) (map[string]string, error) {
	urls := make(map[string]string)
	for _, input := range inputs {
		if !s.isValidURL(input) {
			return nil, errors.New("one or more URLs are invalid")
		}
		shortURL := utils.RandomString(6)
		urls[shortURL] = input
	}

	return s.store.BatchSet(urls)
}

func (s *URLService) GetOriginalURL(input string) (string, bool) {
	return s.store.Get(input)
}
