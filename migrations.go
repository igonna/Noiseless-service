package main

import (
	"database/sql"
	"errors"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func runMigrations() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL must be set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, "./migrations")
}
