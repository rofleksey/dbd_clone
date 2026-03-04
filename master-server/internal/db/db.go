package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init() error {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://dbd:dbd_secret_password@localhost:5432/dbd_clone?sslmode=disable"
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	return nil
}

func RunMigrations() error {
	migrationPath := "migrations/001_init.sql"
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		migrationPath = "/migrations/001_init.sql"
	}

	data, err := os.ReadFile(migrationPath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = DB.Exec(string(data))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
