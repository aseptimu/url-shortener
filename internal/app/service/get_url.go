// Package service содержит бизнес-логику работы с URL.
package service

import "context"

// StatsDTO хранит данные о количестве пользователей и сохраненных url
type StatsDTO struct {
	Urls  int
	Users int
}

type URLDTO struct {
	ShortURL    string
	OriginalURL string
}

// StoreURLGetter описывает методы получения URL из хранилища.
type StoreURLGetter interface {
	Get(ctx context.Context, shortURL string) (originalURL string, err error)
	GetUserURLs(ctx context.Context, userID string) ([]URLDTO, error)
	GetStats(ctx context.Context) (int, int, error)
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
func (s *GetURLService) GetOriginalURL(ctx context.Context, input string) (string, error) {
	return s.store.Get(ctx, input)
}

// GetUserURLs возвращает все URLRecord для данного пользователя.
func (s *GetURLService) GetUserURLs(ctx context.Context, userID string) ([]URLDTO, error) {
	return s.store.GetUserURLs(ctx, userID)
}

// GetStats возвращает количество пользователей и url
func (s *GetURLService) GetStats(ctx context.Context) (StatsDTO, error) {
	urlsCount, usersCount, err := s.store.GetStats(ctx)
	stat := StatsDTO{
		urlsCount,
		usersCount,
	}
	return stat, err
}
