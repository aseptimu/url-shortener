package store

import "sync"

type InMemoryStore struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]string),
	}
}

func (m *InMemoryStore) Get(shortURL string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.data[shortURL]
	return value, exists
}

func (m *InMemoryStore) Set(shortURL, originalURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[shortURL] = originalURL
}
