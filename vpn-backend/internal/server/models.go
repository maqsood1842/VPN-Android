package server

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Server struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Region      string    `json:"region"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	PublicKey   string    `json:"public_key"`
	EndpointIP  string    `json:"endpoint_ip,omitempty"`
	Status      string    `json:"status"`
	LoadPercent int       `json:"load_percent"`
	MaxClients  int       `json:"max_clients"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ServerRepository interface {
	Create(ctx context.Context, s *Server) error
	GetByID(ctx context.Context, id uuid.UUID) (*Server, error)
	ListActive(ctx context.Context) ([]*Server, error)
	UpdateLoad(ctx context.Context, id uuid.UUID, loadPercent int) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}