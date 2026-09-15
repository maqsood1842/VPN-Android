package server

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, s *Server) error {
	query := `
		INSERT INTO servers (id, name, region, host, port, public_key, endpoint_ip, status, load_percent, max_clients, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(ctx, query, s.ID, s.Name, s.Region, s.Host, s.Port, s.PublicKey, s.EndpointIP, s.Status, s.LoadPercent, s.MaxClients, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Server, error) {
	query := `
		SELECT id, name, region, host, port, public_key, endpoint_ip, status, load_percent, max_clients, created_at, updated_at
		FROM servers WHERE id = $1
	`
	s := &Server{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.Region, &s.Host, &s.Port, &s.PublicKey, &s.EndpointIP,
		&s.Status, &s.LoadPercent, &s.MaxClients, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("server not found")
		}
		return nil, err
	}
	return s, nil
}

func (r *Repository) ListActive(ctx context.Context) ([]*Server, error) {
	query := `
		SELECT id, name, region, host, port, public_key, endpoint_ip, status, load_percent, max_clients, created_at, updated_at
		FROM servers WHERE status = 'active' ORDER BY load_percent ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []*Server
	for rows.Next() {
		s := &Server{}
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Region, &s.Host, &s.Port, &s.PublicKey, &s.EndpointIP,
			&s.Status, &s.LoadPercent, &s.MaxClients, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}
	return servers, rows.Err()
}

func (r *Repository) UpdateLoad(ctx context.Context, id uuid.UUID, loadPercent int) error {
	query := `UPDATE servers SET load_percent = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, loadPercent, id)
	return err
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE servers SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}