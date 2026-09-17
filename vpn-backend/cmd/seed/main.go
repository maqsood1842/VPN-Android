package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

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

	// Generate a test WireGuard key pair for the server
	// In production, this would be pre-generated and stored securely
	publicKey := "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg="

	query := `
		INSERT INTO servers (id, name, region, host, port, public_key, endpoint_ip, status, load_percent, max_clients, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT DO NOTHING
	`

	_, err = db.ExecContext(context.Background(), query,
		"New York 1",
		"us-east",
		"nyc1.vpn.example.com",
		51820,
		publicKey,
		"1.2.3.4", // placeholder IP - would be real VPS IP
		"active",
		0,
		100,
	)
	if err != nil {
		log.Fatalf("Failed to seed server: %v", err)
	}

	log.Println("Test server seeded successfully")
}