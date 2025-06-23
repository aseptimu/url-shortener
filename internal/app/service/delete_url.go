// Package service содержит бизнес-логику работы с URL.
package service

import (
	"context"
)

type StoreURLDeleter interface {
	BatchDelete(ctx context.Context, shortURLs []string, userID string) error
}

type URLDeleter interface {
	DeleteURLs(ctx context.Context, shortURLs []string, userID string) error
}

type DeleteURLService struct {
	store StoreURLDeleter
}

func NewURLDeleter(store StoreURLDeleter) *DeleteURLService {
	return &DeleteURLService{store: store}
}
func (s *DeleteURLService) DeleteURLs(ctx context.Context, shortURLs []string, userID string) error {
	return s.store.BatchDelete(ctx, shortURLs, userID)
}
