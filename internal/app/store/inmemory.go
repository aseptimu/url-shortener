// Package store содержит различные реализации хранилищ для URL.
package store

import "sync"

// InMemoryStore хранит URL-пары в памяти,
// используя мапы для прямого и обратного поиска.
type InMemoryStore struct {
	data map[string]string
	rev  map[string]string
	mu   sync.RWMutex
}

// NewStore создаёт и возвращает новый InMemoryStore с инициализированными мапами.
func NewStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]string),
		rev:  make(map[string]string),
	}
}

// Get возвращает оригинальный URL по ключу shortURL
// и булев флаг, указывающий, существует ли такая запись.
func (m *InMemoryStore) Get(shortURL string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, exists := m.data[shortURL]
	return value, exists
}

// Set сохраняет пару shortURL→originalURL.
// Если originalURL уже был сохранён, возвращает существующий короткий URL.
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
