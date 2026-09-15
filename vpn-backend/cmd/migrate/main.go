package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/vpn/backend/pkg/config"
)

func main() {
	cfg := config.Load()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName, cfg.Database.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	migrationsDir := "migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			log.Printf("Warning: failed to read migration %s: %v", file.Name(), err)
			continue
		}

		log.Printf("Running migration: %s", file.Name())
		_, err = db.ExecContext(context.Background(), string(content))
		if err != nil {
			log.Fatalf("Migration %s failed: %v", file.Name(), err)
		}
	}

	log.Println("All migrations completed successfully")
}