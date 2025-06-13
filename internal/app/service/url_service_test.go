package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type stubStore struct {
	setFn   func(ctx context.Context, shortURL, originalURL string) (string, error)
	batchFn func(ctx context.Context, urls map[string]string) (map[string]string, error)
	getFn   func(ctx context.Context, shortURL string) (string, bool)
}

func (s *stubStore) Set(ctx context.Context, shortURL, originalURL string) (string, error) {
	if s.setFn == nil {
		return "", nil
	}
	return s.setFn(ctx, shortURL, originalURL)
}

func (s *stubStore) BatchSet(ctx context.Context, urls map[string]string) (map[string]string, error) {
	if s.batchFn == nil {
		return nil, nil
	}
	return s.batchFn(ctx, urls)
}

func (s *stubStore) Get(ctx context.Context, shortURL string) (string, bool) {
	if s.getFn == nil {
		return "", false
	}
	return s.getFn(ctx, shortURL)
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
	shortURL, err := svc.ShortenURL(context.Background(), input)
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
	_, err := svc.ShortenURL(context.Background(), "invalid-url")
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
	got, err := svc.ShortenURL(context.Background(), input)
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
	got, err := svc.ShortenURLs(context.Background(), inputs)
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
	_, err := svc.ShortenURLs(context.Background(), inputs)
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
	svc := NewURLService(store)
	got, ok := svc.GetOriginalURL(context.Background(), "key")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}
