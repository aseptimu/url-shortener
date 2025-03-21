package service

import (
	"context"
	"errors"
	"net/url"

	"github.com/aseptimu/url-shortener/internal/app/utils"
)

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
}

type Store interface {
	Get(ctx context.Context, shortURL string) (string, bool)
	GetUserURLs(ctx context.Context, userID string) ([]URLRecord, error)
	Set(ctx context.Context, shortURL, originalURL string, userID string) (string, error)
	BatchSet(ctx context.Context, urls map[string]string, userID string) (map[string]string, error)
}

type URLShortener interface {
	ShortenURL(ctx context.Context, input string, userID string) (string, error)
	ShortenURLs(ctx context.Context, inputs []string, userID string) (map[string]string, error)
	GetOriginalURL(ctx context.Context, input string) (string, bool)
	GetUserURLs(ctx context.Context, userID string) ([]URLRecord, error)
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

func (s *URLService) ShortenURL(ctx context.Context, input string, userID string) (string, error) {
	if !s.isValidURL(input) {
		return "", errors.New("invalid URL format")
	}

	shortURL := utils.RandomString(6)
	storeURL, err := s.store.Set(ctx, shortURL, input, userID)
	if err != nil {
		return "", err
	}

	if storeURL != shortURL {
		return storeURL, ErrConflict
	}

	return shortURL, nil
}

func (s *URLService) ShortenURLs(ctx context.Context, inputs []string, userID string) (map[string]string, error) {
	urls := make(map[string]string)
	for _, input := range inputs {
		if !s.isValidURL(input) {
			return nil, errors.New("one or more URLs are invalid")
		}
		shortURL := utils.RandomString(6)
		urls[shortURL] = input
	}

	return s.store.BatchSet(ctx, urls, userID)
}

func (s *URLService) GetOriginalURL(ctx context.Context, input string) (string, bool) {
	return s.store.Get(ctx, input)
}

func (s *URLService) GetUserURLs(ctx context.Context, userID string) ([]URLRecord, error) {
	return s.store.GetUserURLs(ctx, userID)
}
