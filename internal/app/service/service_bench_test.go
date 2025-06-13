package service

import (
	"context"
	"sync"
	"testing"
)

type memStore struct {
	mu   sync.Mutex
	data map[string]string
}

func (m *memStore) Get(_ context.Context, shortURL string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[shortURL]
	return v, ok
}

func (m *memStore) Set(_ context.Context, shortURL, originalURL string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if existing, ok := m.data[shortURL]; ok {
		return existing, nil
	}
	m.data[shortURL] = originalURL
	return originalURL, nil
}

func (m *memStore) BatchSet(_ context.Context, urls map[string]string) (map[string]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for short, orig := range urls {
		if _, exists := m.data[short]; !exists {
			m.data[short] = orig
		}
	}
	return urls, nil
}

func newMemStore() *memStore {
	return &memStore{data: make(map[string]string)}
}

func BenchmarkShortenURL(b *testing.B) {
	svc := NewURLService(newMemStore())
	inputs := make([]string, 100)
	for i := range inputs {
		inputs[i] = "https://test.com/" + string(rune(i))
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ShortenURLs(context.Background(), inputs)
	}
}
