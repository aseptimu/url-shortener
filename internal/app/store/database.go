package store

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"time"
)

type Database struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

const createTableQuery = `
	CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_url TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL UNIQUE
)`

func (db *Database) CreateTables(logger *zap.SugaredLogger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := db.db.ExecContext(ctx, createTableQuery); err != nil {
		logger.Fatalf("Failed to create tables: %v Query:\n%s\n", err, createTableQuery)
	}
}

func NewDB(ps string, logger *zap.SugaredLogger) *Database {
	db, err := sql.Open("pgx", ps)
	if err != nil {
		logger.Panic("failed to connect to database", zap.Error(err))
	}

	return &Database{db, logger}
}

func (db *Database) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := db.db.PingContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) Get(shortURL string) (originalURL string, ok bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row := db.db.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE short_url = $1", shortURL)
	err := row.Scan(&originalURL)
	if err != nil {
		db.logger.Errorw("failed to query url", "shortURL", shortURL, "err", err)
		return "", false
	}

	return originalURL, row != nil
}

const setQuery = `INSERT INTO urls (short_url, original_url) 
         VALUES ($1, $2) 
         ON CONFLICT (original_url) DO NOTHING 
         RETURNING short_url`

func (db *Database) Set(shortURL, originalURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db.logger.Debugw("Attempting to insert URL", "shortURL", shortURL, "originalURL", originalURL)

	err := db.db.QueryRowContext(ctx, setQuery, shortURL, originalURL).Scan(&shortURL)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		db.logger.Errorw("Failed to insert URL", "shortURL", shortURL, "originalURL", originalURL, "err", err)
		return "", err
	}

	if errors.Is(err, sql.ErrNoRows) {
		db.logger.Debugw("URL already exists, fetching short URL from DB", "originalURL", originalURL)

		err = db.db.QueryRowContext(ctx,
			`SELECT short_url FROM urls WHERE original_url = $1`, originalURL).Scan(&shortURL)

		if err != nil {
			db.logger.Errorw("Failed to retrieve existing short URL", "originalURL", originalURL, "err", err)
			return "", err
		}
	}

	db.logger.Debugw("Successfully stored short URL", "shortURL", shortURL, "originalURL", originalURL)
	return shortURL, nil
}

func (db *Database) BatchSet(urls map[string]string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, setQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result := make(map[string]string)

	for shortURL, originalURL := range urls {
		var storedShortURL, storedOriginalURL string
		err = stmt.QueryRowContext(ctx, shortURL, originalURL).Scan(&storedShortURL, &storedOriginalURL)

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			db.logger.Errorw("Failed to insert URL", "shortURL", shortURL, "originalURL", originalURL, "err", err)
			return nil, err
		}

		if errors.Is(err, sql.ErrNoRows) {
			err = tx.QueryRowContext(ctx, `SELECT short_url FROM urls WHERE original_url = $1`, originalURL).Scan(&storedShortURL)
			if err != nil {
				return nil, err
			}
		}

		result[storedShortURL] = storedOriginalURL
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return result, nil
}
