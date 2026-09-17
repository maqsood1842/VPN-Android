package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vpn/backend/pkg/redis"
)

const (
	serverCacheKey     = "servers:active"
	serverCacheTTL     = 30 * time.Second
	serverHealthKey    = "server:health:%s"
	serverHealthTTL    = 60 * time.Second
)

type Repository struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

func NewRepository(db *pgxpool.Pool, redis *redis.Client) *Repository {
	return &Repository{db: db, redis: redis}
}

func (r *Repository) Create(ctx context.Context, s *Server) error {
	query := `
		INSERT INTO servers (id, name, region, host, port, public_key, endpoint_ip, status, load_percent, max_clients, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.Exec(ctx, query, s.ID, s.Name, s.Region, s.Host, s.Port, s.PublicKey, s.EndpointIP, s.Status, s.LoadPercent, s.MaxClients, s.CreatedAt, s.UpdatedAt)
	if err == nil && r.redis != nil {
		r.invalidateCache(ctx)
	}
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
	
	// Enrich with cached health data
	if r.redis != nil {
		r.enrichFromCache(ctx, s)
	}
	return s, nil
}

func (r *Repository) ListActive(ctx context.Context) ([]*Server, error) {
	// Try cache first
	if r.redis != nil {
		cached, err := r.getFromCache(ctx)
		if err == nil && len(cached) > 0 {
			return cached, nil
		}
	}

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

	// Enrich with cached health data
	if r.redis != nil {
		for _, s := range servers {
			r.enrichFromCache(ctx, s)
		}
		// Cache the enriched list
		r.setCache(ctx, servers)
	}

	return servers, rows.Err()
}

func (r *Repository) UpdateLoad(ctx context.Context, id uuid.UUID, loadPercent int) error {
	query := `UPDATE servers SET load_percent = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, loadPercent, id)
	if err == nil && r.redis != nil {
		r.invalidateCache(ctx)
		// Update health cache
		r.updateHealthCache(ctx, id, loadPercent, "active")
	}
	return err
}

func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE servers SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	if err == nil && r.redis != nil {
		r.invalidateCache(ctx)
		r.updateHealthCache(ctx, id, 0, status)
	}
	return err
}

func (r *Repository) UpdateHealth(ctx context.Context, id uuid.UUID, loadPercent int, status string, peerCount int) error {
	query := `UPDATE servers SET load_percent = $1, status = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(ctx, query, loadPercent, status, id)
	if err == nil && r.redis != nil {
		r.invalidateCache(ctx)
		r.updateHealthCache(ctx, id, loadPercent, status)
	}
	return err
}

func (r *Repository) getFromCache(ctx context.Context) ([]*Server, error) {
	data, err := r.redis.Get(ctx, serverCacheKey).Bytes()
	if err != nil {
		return nil, err
	}
	var servers []*Server
	err = json.Unmarshal(data, &servers)
	return servers, err
}

func (r *Repository) setCache(ctx context.Context, servers []*Server) {
	data, err := json.Marshal(servers)
	if err != nil {
		return
	}
	r.redis.Set(ctx, serverCacheKey, data, serverCacheTTL)
}

func (r *Repository) invalidateCache(ctx context.Context) {
	r.redis.Del(ctx, serverCacheKey)
}

func (r *Repository) enrichFromCache(ctx context.Context, s *Server) {
	key := fmt.Sprintf(serverHealthKey, s.ID.String())
	data, err := r.redis.Get(ctx, key).Bytes()
	if err != nil {
		return
	}
	var health ServerHealth
	if json.Unmarshal(data, &health) == nil {
		s.LoadPercent = health.LoadPercent
		s.Status = health.Status
	}
}

func (r *Repository) updateHealthCache(ctx context.Context, id uuid.UUID, loadPercent int, status string) {
	key := fmt.Sprintf(serverHealthKey, id.String())
	health := ServerHealth{
		ServerID:    id.String(),
		LoadPercent: loadPercent,
		Status:      status,
		UpdatedAt:   time.Now(),
	}
	data, _ := json.Marshal(health)
	r.redis.Set(ctx, key, data, serverHealthTTL)
}

type ServerHealth struct {
	ServerID    string    `json:"server_id"`
	LoadPercent int       `json:"load_percent"`
	Status      string    `json:"status"`
	UpdatedAt   time.Time `json:"updated_at"`
}