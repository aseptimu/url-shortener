// Package service содержит бизнес-логику работы с URL.
package service

import (
	"context"
)

// StoreURLDeleter описывает метод пакетного удаления URL из хранилища.
type StoreURLDeleter interface {
	BatchDelete(ctx context.Context, shortURLs []string, userID string) error
}

// URLDeleter предоставляет возможность удаления URL.
type URLDeleter interface {
	DeleteURLs(ctx context.Context, shortURLs []string, userID string) error
}

// DeleteURLService реализует URLDeleter через StoreURLDeleter.
type DeleteURLService struct {
	store StoreURLDeleter
}

// NewURLDeleter создаёт новый сервис для удаления URL на основе переданного хранилища.
func NewURLDeleter(store StoreURLDeleter) *DeleteURLService {
	return &DeleteURLService{store: store}
}

// DeleteURLs вызывает BatchDelete у внутреннего хранилища для удаления списка shortURLs.
func (s *DeleteURLService) DeleteURLs(ctx context.Context, shortURLs []string, userID string) error {
	return s.store.BatchDelete(ctx, shortURLs, userID)
}
