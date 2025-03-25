package store

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"log"
	"os"
	"sync"
)

type FileStore struct {
	mu       sync.RWMutex
	filePath string
	data     map[string]service.URLRecord
}

func NewFileStore(filePath string) *FileStore {
	store := &FileStore{
		filePath: filePath,
		data:     make(map[string]service.URLRecord),
	}
	store.loadFromFile()
	return store
}

func (fs *FileStore) loadFromFile() {
	file, err := os.Open(fs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return // Если файла нет, продолжаем
		}
		log.Panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record service.URLRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err == nil {
			fs.data[record.ShortURL] = record
		}
	}
}

func (fs *FileStore) saveToFile(record service.URLRecord) {
	file, err := os.OpenFile(fs.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	jsonData, _ := json.Marshal(record)
	file.Write(jsonData)
	file.Write([]byte("\n"))
}

func (fs *FileStore) rewriteFile() error {
	file, err := os.OpenFile(fs.filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, record := range fs.data {
		jsonData, _ := json.Marshal(record)
		writer.Write(jsonData)
		writer.WriteString("\n")
	}
	return writer.Flush()
}

func (fs *FileStore) Get(_ context.Context, shortURL string) (string, bool, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	record, exists := fs.data[shortURL]
	if !exists || record.DeletedFlag {
		return "", false, false
	}
	return record.OriginalURL, true, record.DeletedFlag
}

func (fs *FileStore) GetUserURLs(_ context.Context, userID string) ([]service.URLRecord, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var results []service.URLRecord
	for _, record := range fs.data {
		if record.UserID == userID && !record.DeletedFlag {
			results = append(results, record)
		}
	}
	return results, nil
}

func (fs *FileStore) Set(_ context.Context, shortURL, originalURL, userID string) (string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for _, record := range fs.data {
		if record.OriginalURL == originalURL {
			return record.ShortURL, nil
		}
	}

	newRecord := service.URLRecord{
		UUID:        shortURL,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
		DeletedFlag: false,
	}
	fs.data[shortURL] = newRecord
	fs.saveToFile(newRecord)

	return shortURL, nil
}

func (fs *FileStore) BatchSet(_ context.Context, urls map[string]string, userID string) (map[string]string, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	file, err := os.OpenFile(fs.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	shortenedURLs := make(map[string]string)

	for shortURL, originalURL := range urls {
		found := false
		for _, record := range fs.data {
			if record.OriginalURL == originalURL {
				shortenedURLs[originalURL] = record.ShortURL
				found = true
				break
			}
		}
		if found {
			continue
		}

		newRecord := service.URLRecord{
			UUID:        shortURL,
			ShortURL:    shortURL,
			OriginalURL: originalURL,
			UserID:      userID,
			DeletedFlag: false,
		}
		fs.data[shortURL] = newRecord
		shortenedURLs[originalURL] = shortURL

		jsonData, _ := json.Marshal(newRecord)
		file.Write(jsonData)
		file.Write([]byte("\n"))
	}

	return shortenedURLs, nil
}

func (fs *FileStore) BatchDelete(_ context.Context, shortURLs []string, userID string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for _, shortURL := range shortURLs {
		record, exists := fs.data[shortURL]
		if exists && record.UserID == userID {
			record.DeletedFlag = true
			fs.data[shortURL] = record
		}
	}

	return fs.rewriteFile()
}
