package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type stubStore struct {
	setFn         func(ctx context.Context, shortURL, originalURL string) (string, error)
	batchFn       func(ctx context.Context, urls map[string]string) (map[string]string, error)
	getFn         func(ctx context.Context, shortURL string) (string, bool)
	getUserURLsFn func(ctx context.Context, userID string) ([]URLDTO, error)
}

func (s *stubStore) GetUserURLs(ctx context.Context, userID string) ([]URLDTO, error) {
	if s.getUserURLsFn != nil {
		return s.getUserURLsFn(ctx, userID)
	}
	return nil, nil
}

func (s *stubStore) GetStats(_ context.Context) (int, int, error) {
	return 0, 0, nil
}

func (s *stubStore) Set(ctx context.Context, shortURL, originalURL string, _ string) (string, error) {
	if s.setFn == nil {
		return "", nil
	}
	return s.setFn(ctx, shortURL, originalURL)
}

func (s *stubStore) BatchSet(ctx context.Context, urls map[string]string, _ string) (map[string]string, error) {
	if s.batchFn == nil {
		return nil, nil
	}
	return s.batchFn(ctx, urls)
}

func (s *stubStore) Get(ctx context.Context, shortURL string) (originalURL string, err error) {
	if s.getFn == nil {
		return "", nil
	}
	url, _ := s.getFn(ctx, shortURL)
	return url, ErrURLDeleted
}

func TestShortenURL_Success(t *testing.T) {
	var called bool
	var passedURL string
	store := &stubStore{
		setFn: func(ctx context.Context, shortURL, originalURL string) (string, error) {
			called = true
			passedURL = originalURL
			return shortURL, nil
		},
	}
	svc := NewURLService(store)
	input := "https://example.com"
	shortURL, err := svc.ShortenURL(context.Background(), input, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Error("expected Set to be called")
	}
	if shortURL == "" {
		t.Error("expected non-empty shortURL")
	}
	if passedURL != input {
		t.Errorf("expected Set called with originalURL %q, got %q", input, passedURL)
	}
}

func TestShortenURL_InvalidURL(t *testing.T) {
	svc := NewURLService(&stubStore{})
	_, err := svc.ShortenURL(context.Background(), "invalid-url", "")
	if err == nil || err.Error() != "invalid URL format" {
		t.Fatalf("expected invalid URL format error, got %v", err)
	}
}

func TestShortenURL_Conflict(t *testing.T) {
	expected := "existingKey"
	store := &stubStore{
		setFn: func(ctx context.Context, shortURL, originalURL string) (string, error) {
			return expected, nil
		},
	}
	svc := NewURLService(store)
	input := "https://example.com"
	got, err := svc.ShortenURL(context.Background(), input, "")
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if got != expected {
		t.Errorf("expected shortURL %q, got %q", expected, got)
	}
}

func TestShortenURLs_Success(t *testing.T) {
	inputs := []string{"https://a", "https://b"}
	want := map[string]string{"shortA": "https://a", "shortB": "https://b"}
	store := &stubStore{
		batchFn: func(ctx context.Context, urls map[string]string) (map[string]string, error) {
			return want, nil
		},
	}
	svc := NewURLService(store)
	got, err := svc.ShortenURLs(context.Background(), inputs, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestShortenURLs_InvalidURL(t *testing.T) {
	svc := NewURLService(&stubStore{})
	inputs := []string{"invalid", "https://example.com"}
	_, err := svc.ShortenURLs(context.Background(), inputs, "")
	if err == nil || err.Error() != "one or more URLs are invalid" {
		t.Fatalf("expected invalid URL error, got %v", err)
	}
}

func TestGetOriginalURL(t *testing.T) {
	expected := "origURL"
	store := &stubStore{
		getFn: func(ctx context.Context, shortURL string) (string, bool) {
			return expected, true
		},
	}
	svc := NewGetURLService(store)
	got, err := svc.GetOriginalURL(context.Background(), "key")
	if !errors.Is(err, ErrURLDeleted) {
		t.Fatal(err)
	}
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
func TestGetUserURLs_Success(t *testing.T) {
	dummy := []URLDTO{
		{ShortURL: "s1", OriginalURL: "o1"},
		{ShortURL: "s2", OriginalURL: "o2"},
	}

	store := &stubStore{
		getUserURLsFn: func(ctx context.Context, userID string) ([]URLDTO, error) {
			if userID != "u1" {
				t.Errorf("expected userID u1, got %q", userID)
			}
			return dummy, nil
		},
	}
	svc := NewGetURLService(store)

	got, err := svc.GetUserURLs(context.Background(), "u1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !reflect.DeepEqual(got, dummy) {
		t.Errorf("expected %v, got %v", dummy, got)
	}
}

func TestGetUserURLs_Error(t *testing.T) {
	expectedErr := errors.New("something went wrong")
	store := &stubStore{
		getUserURLsFn: func(ctx context.Context, userID string) ([]URLDTO, error) {
			return nil, expectedErr
		},
	}
	svc := NewGetURLService(store)

	got, err := svc.GetUserURLs(context.Background(), "any")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
	if got != nil {
		t.Errorf("expected nil slice on error, got %v", got)
	}
}
