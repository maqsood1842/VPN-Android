package wireguard

import (
	"time"

	"github.com/google/uuid"
)

type Peer struct {
	ID                 uuid.UUID  `json:"id"`
	UserID             uuid.UUID  `json:"user_id"`
	ServerID           uuid.UUID  `json:"server_id"`
	PrivateKey         string     `json:"private_key"`
	PublicKey          string     `json:"public_key"`
	PresharedKey       string     `json:"preshared_key,omitempty"`
	AllowedIPs         []string   `json:"allowed_ips"`
	DNSServers         []string   `json:"dns_servers"`
	EndpointIP         string     `json:"endpoint_ip"`
	PersistentKeepalive int       `json:"persistent_keepalive"`
	Status             string     `json:"status"`
	LastHandshake      *time.Time `json:"last_handshake,omitempty"`
	BytesReceived      int64      `json:"bytes_received"`
	BytesSent          int64      `json:"bytes_sent"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type PeerConfig struct {
	Interface struct {
		PrivateKey string
		Address    []string
		DNS        []string
	}
	Peer struct {
		PublicKey           string
		PresharedKey        string
		AllowedIPs          []string
		Endpoint            string
		PersistentKeepalive int
	}
}

type PeerRepository interface {
	Create(ctx context.Context, p *Peer) error
	GetByUserAndServer(ctx context.Context, userID, serverID uuid.UUID) (*Peer, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*Peer, error)
	Update(ctx context.Context, p *Peer) error
}

type ConfigGenerator interface {
	GenerateConfig(ctx context.Context, userID, serverID uuid.UUID) (*PeerConfig, error)
}