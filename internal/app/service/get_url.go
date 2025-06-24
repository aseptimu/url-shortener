// Package service содержит бизнес-логику работы с URL.
package service

import "context"

// StoreURLGetter описывает методы получения URL из хранилища.
type StoreURLGetter interface {
	Get(ctx context.Context, shortURL string) (originalURL string, deleted bool, exists bool)
	GetUserURLs(ctx context.Context, userID string) ([]URLRecord, error)
}

// URLGetter предоставляет методы получения URL для клиентского кода.
type URLGetter interface {
	GetOriginalURL(ctx context.Context, input string) (string, bool, bool)
	GetUserURLs(ctx context.Context, userID string) ([]URLRecord, error)
}

// GetURLService реализует URLGetter через StoreURLGetter.
type GetURLService struct {
	store StoreURLGetter
}

// NewGetURLService создаёт новый GetURLService на основе переданного хранилища.
func NewGetURLService(store StoreURLGetter) *GetURLService {
	return &GetURLService{store: store}
}

// GetOriginalURL возвращает оригинальный URL, exists и deleted.
func (s *GetURLService) GetOriginalURL(ctx context.Context, input string) (string, bool, bool) {
	return s.store.Get(ctx, input)
}

// GetUserURLs возвращает все URLRecord для данного пользователя.
func (s *GetURLService) GetUserURLs(ctx context.Context, userID string) ([]URLRecord, error) {
	return s.store.GetUserURLs(ctx, userID)
}
