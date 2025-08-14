// Package service содержит бизнес-логику работы с URL.
package service

import (
	"context"
	"errors"
	"net/url"

	"github.com/aseptimu/url-shortener/internal/app/utils"
)

// Store объединяет интерфейсы для получения, создания и удаления URL.
type Store interface {
	StoreURLGetter
	StoreURLSetter
	StoreURLDeleter
}

// StoreURLSetter описывает методы сохранения одного или нескольких URL.
type StoreURLSetter interface {
	Set(ctx context.Context, shortURL, originalURL, userID string) (string, error)
	BatchSet(ctx context.Context, urls map[string]string, userID string) (map[string]string, error)
}

// URLShortener предоставляет методы для сокращения одного или нескольких URL.
type URLShortener interface {
	ShortenURL(ctx context.Context, input string, userID string) (string, error)
	ShortenURLs(ctx context.Context, inputs []string, userID string) (map[string]string, error)
}

// URLService реализует URLShortener через StoreURLSetter.
type URLService struct {
	store StoreURLSetter
}

// NewURLService создаёт новый URLService.
func NewURLService(store StoreURLSetter) *URLService {
	return &URLService{store: store}
}

func (s *URLService) isValidURL(input string) bool {
	parsedURI, err := url.ParseRequestURI(input)
	return err == nil && parsedURI.Scheme != "" && parsedURI.Host != ""
}

// ErrConflict возвращается, если оригинальный URL уже существует.
var ErrConflict = errors.New("URL already exists")

// ShortenURL создаёт короткий URL для данного входа или возвращает ErrConflict
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

// ShortenURLs создаёт короткие ссылки для нескольких URL, возвращая карту shortURL→originalURL.
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
