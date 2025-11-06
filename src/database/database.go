package database

import (
	"github.com/jmoiron/sqlx"

	"digitalUniversity/config"
)

func OpenDB(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", cfg.URI)
	if err != nil {
		return nil, err
	}

	return db, nil
}
