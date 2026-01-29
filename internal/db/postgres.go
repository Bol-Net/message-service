package db

import (
	"database/sql"
	"messaging-service/internal/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init(cfg *config.Config) error {
	var err error
	DB, err = sql.Open("postgres", cfg.PostgresDSN)
	if err != nil {
		return err
	}
	return DB.Ping()
}
