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

const createTableQuery = `CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_url TEXT NOT NULL,
    original_url TEXT NOT NULL
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
	}

	return originalURL, row != nil
}
func (db *Database) Set(shortURL, originalURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.db.QueryRowContext(ctx,
		`INSERT INTO urls (short_url, original_url) 
         VALUES ($1, $2) 
         ON CONFLICT (original_url) DO NOTHING 
         RETURNING short_url`,
		shortURL, originalURL).Scan(&shortURL)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		db.logger.Errorw("failed to set url", "short_url", shortURL, "original_url", originalURL, "err", err)
		return "", err
	}

	if errors.Is(err, sql.ErrNoRows) {
		db.logger.Debugw("URL already exists", "original_url", originalURL)
		err = db.db.QueryRowContext(ctx,
			`SELECT short_url FROM urls WHERE original_url = $1`, originalURL).Scan(&shortURL)
		if err != nil {
			return "", err
		}
	}
	return shortURL, err
}
