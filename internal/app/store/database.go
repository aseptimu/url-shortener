package store

import (
	"context"
	"database/sql"
	"errors"
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"go.uber.org/zap"
)

type Database struct {
	dbpool *pgxpool.Pool
	logger *zap.SugaredLogger
}

const CreateTableQuery = `
	CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_url TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL UNIQUE
)`

func (db *Database) CreateTables(logger *zap.SugaredLogger) {
	ctx, cancel := context.WithTimeout(context.Background(), config.DBTimeout)
	defer cancel()
	if _, err := db.dbpool.Exec(ctx, CreateTableQuery); err != nil {
		logger.Fatalf("Failed to create tables: %v Query:\n%s\n", err, CreateTableQuery)
	}
}

func NewDB(ps string, logger *zap.SugaredLogger) *Database {
	dbpool, err := pgxpool.New(context.Background(), ps)
	if err != nil {
		logger.Panic("failed to connect to database", zap.Error(err))
	}

	return &Database{dbpool, logger}
}

func (db *Database) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, config.DBTimeout)
	defer cancel()
	err := db.dbpool.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}

const GetURLQuery = "SELECT original_url FROM urls WHERE short_url = $1"

func (db *Database) Get(ctx context.Context, shortURL string) (originalURL string, ok bool) {
	ctx, cancel := context.WithTimeout(ctx, config.DBTimeout)
	defer cancel()

	row := db.dbpool.QueryRow(ctx, GetURLQuery, shortURL)
	err := row.Scan(&originalURL)
	if err != nil {
		db.logger.Errorw("failed to query url", "shortURL", shortURL, "err", err)
		return "", false
	}

	return originalURL, row != nil
}

const SetURLQuery = `INSERT INTO urls (short_url, original_url) 
         VALUES ($1, $2) 
         ON CONFLICT (original_url) DO NOTHING 
         RETURNING short_url`

const GetExistingURLQuery = "SELECT short_url FROM urls WHERE original_url = $1"

func (db *Database) Set(ctx context.Context, shortURL, originalURL string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, config.DBTimeout)
	defer cancel()

	db.logger.Debugw("Attempting to insert URL", "shortURL", shortURL, "originalURL", originalURL)

	err := db.dbpool.QueryRow(ctx, SetURLQuery, shortURL, originalURL).Scan(&shortURL)

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

func (db *Database) BatchSet(ctx context.Context, urls map[string]string) (map[string]string, error) {
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
		err = tx.QueryRow(ctx, SetURLQuery, shortURL, originalURL).Scan(&storedShortURL)

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
