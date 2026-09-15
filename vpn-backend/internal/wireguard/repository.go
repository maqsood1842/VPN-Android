package wireguard

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

func (r *Repository) Create(ctx context.Context, p *Peer) error {
	query := `
		INSERT INTO wg_peers (id, user_id, server_id, private_key, public_key, preshared_key, allowed_ips, dns_servers, endpoint_ip, persistent_keepalive, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.Exec(ctx, query,
		p.ID, p.UserID, p.ServerID, p.PrivateKey, p.PublicKey, p.PresharedKey,
		p.AllowedIPs, p.DNSServers, p.EndpointIP, p.PersistentKeepalive, p.Status,
		p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *Repository) GetByUserAndServer(ctx context.Context, userID, serverID uuid.UUID) (*Peer, error) {
	query := `
		SELECT id, user_id, server_id, private_key, public_key, preshared_key, allowed_ips, dns_servers, endpoint_ip, persistent_keepalive, status, last_handshake, bytes_received, bytes_sent, created_at, updated_at
		FROM wg_peers WHERE user_id = $1 AND server_id = $2
	`
	p := &Peer{}
	err := r.db.QueryRow(ctx, query, userID, serverID).Scan(
		&p.ID, &p.UserID, &p.ServerID, &p.PrivateKey, &p.PublicKey, &p.PresharedKey,
		&p.AllowedIPs, &p.DNSServers, &p.EndpointIP, &p.PersistentKeepalive, &p.Status,
		&p.LastHandshake, &p.BytesReceived, &p.BytesSent, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("peer not found")
		}
		return nil, err
	}
	return p, nil
}

func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Peer, error) {
	query := `
		SELECT id, user_id, server_id, private_key, public_key, preshared_key, allowed_ips, dns_servers, endpoint_ip, persistent_keepalive, status, last_handshake, bytes_received, bytes_sent, created_at, updated_at
		FROM wg_peers WHERE user_id = $1
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peers []*Peer
	for rows.Next() {
		p := &Peer{}
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.ServerID, &p.PrivateKey, &p.PublicKey, &p.PresharedKey,
			&p.AllowedIPs, &p.DNSServers, &p.EndpointIP, &p.PersistentKeepalive, &p.Status,
			&p.LastHandshake, &p.BytesReceived, &p.BytesSent, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	return peers, rows.Err()
}

func (r *Repository) Update(ctx context.Context, p *Peer) error {
	query := `
		UPDATE wg_peers SET
			private_key = $2, public_key = $3, preshared_key = $4,
			allowed_ips = $5, dns_servers = $6, endpoint_ip = $7,
			persistent_keepalive = $8, status = $9, last_handshake = $10,
			bytes_received = $11, bytes_sent = $12, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		p.ID, p.PrivateKey, p.PublicKey, p.PresharedKey,
		p.AllowedIPs, p.DNSServers, p.EndpointIP, p.PersistentKeepalive,
		p.Status, p.LastHandshake, p.BytesReceived, p.BytesSent,
	)
	return err
}