package store

import "sync"

type InMemoryStore struct {
	data map[string]string
	rev  map[string]string
	mu   sync.RWMutex
}

func NewStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]string),
		rev:  make(map[string]string),
	}
}

func (m *InMemoryStore) Get(shortURL string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.data[shortURL]
	return value, exists
}

func (m *InMemoryStore) Set(shortURL, originalURL string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existingShort, found := m.rev[originalURL]; found {
		return existingShort, nil
	}

	m.data[shortURL] = originalURL
	m.rev[originalURL] = shortURL
	return shortURL, nil
}
