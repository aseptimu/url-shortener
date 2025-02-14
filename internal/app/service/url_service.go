package service

import (
	"errors"
	"net/url"

	"github.com/aseptimu/url-shortener/internal/app/store"
	"github.com/aseptimu/url-shortener/internal/app/utils"
)

type URLShortener interface {
	ShortenURL(input string) (string, error)
	GetOriginalURL(input string) (string, bool)
}

type URLService struct {
	store store.Store
}

func NewURLService(store store.Store) *URLService {
	return &URLService{store: store}
}

func (s *URLService) isValidURL(input string) bool {
	parsedURI, err := url.ParseRequestURI(input)
	return err == nil && parsedURI.Scheme != "" && parsedURI.Host != ""
}

func (s *URLService) ShortenURL(input string) (string, error) {
	if !s.isValidURL(input) {
		return "", errors.New("invalid URL format")
	}

	shortURL := utils.RandomString(6)

	s.store.Set(shortURL, input)

	return shortURL, nil
}

func (s *URLService) GetOriginalURL(input string) (string, bool) {
	return s.store.Get(input)
}
