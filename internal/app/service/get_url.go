package service

import (
	"context"
	"net/url"
)

type StoreURLGetter interface {
	Get(ctx context.Context, shortURL string) (originalURL string, deleted bool, exists bool)
	GetUserURLs(ctx context.Context, userID string) ([]URLRecord, error)
}

type URLGetter interface {
	GetOriginalURL(ctx context.Context, input string) (string, bool, bool)
	GetUserURLs(ctx context.Context, userID string) ([]URLRecord, error)
}

type GetURLService struct {
	store StoreURLGetter
}

func NewGetURLService(store StoreURLGetter) *GetURLService {
	return &GetURLService{store: store}
}

func (s *GetURLService) isValidURL(input string) bool {
	parsedURI, err := url.ParseRequestURI(input)
	return err == nil && parsedURI.Scheme != "" && parsedURI.Host != ""
}

func (s *GetURLService) GetOriginalURL(ctx context.Context, input string) (string, bool, bool) {
	return s.store.Get(ctx, input)
}

func (s *GetURLService) GetUserURLs(ctx context.Context, userID string) ([]URLRecord, error) {
	return s.store.GetUserURLs(ctx, userID)
}
