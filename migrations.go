package main

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func runMigrations() error {
	if shouldSkipMigrations() {
		log.Println("skipping goose migrations (SKIP_MIGRATIONS=1)")
		return nil
	}

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

func shouldSkipMigrations() bool {
	val := strings.TrimSpace(os.Getenv("SKIP_MIGRATIONS"))
	if val == "" {
		val = strings.TrimSpace(os.Getenv("DISABLE_GOOSE"))
	}
	switch strings.ToLower(val) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
