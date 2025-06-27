// Package store содержит функции для миграции базы данных.
package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

// MigrateDB подключается к базе, настраивает параметры соединения,
// а затем выполняет вложенные миграции из каталога ./migrations.
// В случае ошибки открытия, инициализации драйвера или выполнения миграций
// возвращает соответствующую ошибку.
func MigrateDB(ps string, logger *zap.SugaredLogger) error {
	db, err := sql.Open("pgx", ps)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Infof("Migration executed successfully")
	return nil
}
