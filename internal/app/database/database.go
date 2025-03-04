package database

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func NewDB(ps string, logger *zap.Logger) *sql.DB {
	db, err := sql.Open("pgx", ps)
	if err != nil {
		logger.Panic("failed to connect to database", zap.Error(err))
	}

	return db
}
