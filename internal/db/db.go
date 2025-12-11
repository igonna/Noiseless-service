package db

import (
	"database/sql"
	"errors"
	"os"
	"strings"
	"sync"

	_ "github.com/lib/pq"
)

var (
	dbOnce sync.Once
	appDB  *sql.DB
	dbErr  error
)

// Init lazily opens a shared *sql.DB using DATABASE_URL.
func Init() (*sql.DB, error) {
	dbOnce.Do(func() {
		dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
		if dsn == "" {
			dbErr = errors.New("DATABASE_URL must be set")
			return
		}

		conn, err := sql.Open("postgres", dsn)
		if err != nil {
			dbErr = err
			return
		}

		if err := conn.Ping(); err != nil {
			conn.Close()
			dbErr = err
			return
		}

		appDB = conn
	})

	return appDB, dbErr
}

func Get() (*sql.DB, error) {
	return Init()
}
