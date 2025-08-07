// Package store содержит реализацию хранилища URL в PostgreSQL.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.uber.org/zap"
)

// Database управляет подключением к PostgreSQL и логированием.
type Database struct {
	dbpool *pgxpool.Pool
	logger *zap.SugaredLogger
}

// NewDB создаёт и возвращает новый Database,
// устанавливая соединение по строке ps и используя logger для логирования.
func NewDB(ps string, logger *zap.SugaredLogger) *Database {
	dbpool, err := pgxpool.New(context.Background(), ps)
	if err != nil {
		logger.Panic("failed to connect to database", zap.Error(err))
	}

	return &Database{dbpool, logger}
}

// Ping проверяет доступность базы данных в пределах таймаута config.DBTimeout.
func (db *Database) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, config.DBTimeout)
	defer cancel()
	if db == nil {
		return errors.New("database is not initialized")
	}
	err := db.dbpool.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}

// GetURLQuery содержит SQL-запрос для получения оригинального URL и флага удаления.
const GetURLQuery = "SELECT original_url, is_deleted FROM urls WHERE short_url = $1"

// Get возвращает originalURL, признак существования и признак удаления для shortURL.
func (db *Database) Get(ctx context.Context, shortURL string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, config.DBTimeout)
	defer cancel()

	var originalURL string
	var deleted bool

	row := db.dbpool.QueryRow(ctx, GetURLQuery, shortURL)
	err := row.Scan(&originalURL, &deleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return "", service.ErrURLNotFound
		}
		db.logger.Errorw("failed to query url", "shortURL", shortURL, "err", err)
		return "", fmt.Errorf("database error: %w", err)
	}

	if deleted {
		return "", service.ErrURLDeleted
	}

	return originalURL, nil
}

// GetURLsByUserID содержит SQL-запрос для получения всех URL пользователя.
const GetURLsByUserID = "SELECT short_url, original_url FROM urls WHERE user_id = $1"

// GetUserURLs возвращает список service.URLRecord для заданного userID.
func (db *Database) GetUserURLs(ctx context.Context, userID string) ([]service.URLDTO, error) {
	rows, err := db.dbpool.Query(ctx, GetURLsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []service.URLDTO
	for rows.Next() {
		var rec service.URLDTO
		if err := rows.Scan(&rec.ShortURL, &rec.OriginalURL); err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	return results, nil
}

// SetURLQuery содержит SQL-запрос для вставки новой записи или пропуска при конфликте.
const SetURLQuery = `INSERT INTO urls (short_url, original_url, user_id) 
         VALUES ($1, $2, $3) 
         ON CONFLICT (original_url) DO NOTHING 
         RETURNING short_url`

// GetExistingURLQuery содержит SQL-запрос для получения существующего short_url по original_url.
const GetExistingURLQuery = "SELECT short_url FROM urls WHERE original_url = $1"

// Set сохраняет пару shortURL→originalURL и возвращает фактический ключ.
// В случае конфликта возвращает уже существующий shortURL.
func (db *Database) Set(ctx context.Context, shortURL, originalURL string, userID string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, config.DBTimeout)
	defer cancel()

	db.logger.Debugw("Attempting to insert URL", "shortURL", shortURL, "originalURL", originalURL)

	err := db.dbpool.QueryRow(ctx, SetURLQuery, shortURL, originalURL, userID).Scan(&shortURL)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		db.logger.Errorw("Failed to insert URL", "shortURL", shortURL, "originalURL", originalURL, "err", err)
		return "", err
	}

	if errors.Is(err, sql.ErrNoRows) {
		db.logger.Debugw("URL already exists, fetching short URL from DB", "originalURL", originalURL)

		err = db.dbpool.QueryRow(ctx, GetExistingURLQuery, originalURL).Scan(&shortURL)

		if err != nil {
			db.logger.Errorw("Failed to retrieve existing short URL", "originalURL", originalURL, "err", err)
			return "", err
		}
	}

	db.logger.Debugw("Successfully stored short URL", "shortURL", shortURL, "originalURL", originalURL)
	return shortURL, nil
}

// BatchSet сохраняет несколько URL в рамках одной транзакции и возвращает мапу shortURL→originalURL.
func (db *Database) BatchSet(ctx context.Context, urls map[string]string, userID string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, config.DBTimeout)
	defer cancel()

	tx, err := db.dbpool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	result := make(map[string]string)

	for shortURL, originalURL := range urls {
		var storedShortURL string
		err = tx.QueryRow(ctx, SetURLQuery, shortURL, originalURL, userID).Scan(&storedShortURL)

		if err != nil && err != pgx.ErrNoRows {
			db.logger.Errorw("Failed to insert URL", "shortURL", shortURL, "originalURL", originalURL, "err", err)
			return nil, err
		}

		// Если вставка не сработала (конфликт), получаем уже существующую короткую ссылку
		if err == pgx.ErrNoRows || storedShortURL == "" {
			err = tx.QueryRow(ctx, GetExistingURLQuery, originalURL).Scan(&storedShortURL)
			if err != nil {
				db.logger.Errorw("Failed to retrieve existing short URL", "originalURL", originalURL, "err", err)
				return nil, err
			}
		}

		result[storedShortURL] = originalURL
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return result, nil
}

// BatchDeleteQuery содержит SQL-запрос для пометки URL как удалённых.
const BatchDeleteQuery = "UPDATE urls SET is_deleted = TRUE WHERE short_url = ANY($1) AND user_id = $2"

// BatchDelete помечает указанные shortURLs как удалённые для заданного userID.
func (db *Database) BatchDelete(ctx context.Context, shortURLs []string, userID string) error {
	cmdTag, err := db.dbpool.Exec(ctx, BatchDeleteQuery, shortURLs, userID)
	if err != nil {
		db.logger.Errorw("Failed to batch delete URLs", "error", err)
		return err
	}

	db.logger.Debugw("Batch delete completed", "rowsAffected", cmdTag.RowsAffected())
	return nil
}

// GetStatQuery возвращает количество
const GetStatQuery = "SELECT COUNT(short_url), COUNT(DISTINCT(user_id)) FROM urls;"

// GetStats возвращает количество пользователей и url
func (db *Database) GetStats(ctx context.Context) (users, urls int, err error) {
	row := db.dbpool.QueryRow(ctx, GetStatQuery)

	err = row.Scan(&users, &urls)
	if err != nil {
		db.logger.Errorw("Failed get stats", "error", err)
		return 0, 0, err
	}

	return users, urls, nil
}
