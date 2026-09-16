package server

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/vpn/backend/internal/wireguard"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type HealthChecker struct {
	serverRepo *Repository
	wgClient   *wgctrl.Client
	interval   time.Duration
	stopCh     chan struct{}
}

func NewHealthChecker(serverRepo *Repository, interval time.Duration) (*HealthChecker, error) {
	var client *wgctrl.Client
	var err error
	if client, err = wgctrl.New(); err != nil {
		log.Printf("Warning: wgctrl not available for health checks: %v", err)
	}

	return &HealthChecker{
		serverRepo: serverRepo,
		wgClient:   client,
		interval:   interval,
		stopCh:     make(chan struct{}),
	}, nil
}

func (h *HealthChecker) Start(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	// Run initial check
	h.checkAllServers(ctx)

	for {
		select {
		case <-ticker.C:
			h.checkAllServers(ctx)
		case <-h.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (h *HealthChecker) Stop() {
	close(h.stopCh)
}

func (h *HealthChecker) checkAllServers(ctx context.Context) {
	servers, err := h.serverRepo.ListActive(ctx)
	if err != nil {
		log.Printf("Health check: failed to list servers: %v", err)
		return
	}

	for _, s := range servers {
		h.checkServer(ctx, s)
	}
}

func (h *HealthChecker) checkServer(ctx context.Context, s *Server) {
	healthy := true
	loadPercent := 0
	peerCount := 0

	// Check if WireGuard interface is up and get peer count
	if h.wgClient != nil {
		device, err := h.wgClient.Device("wg0")
		if err != nil {
			healthy = false
		} else {
			peerCount = len(device.Peers)
			// Calculate load based on max clients
			if s.MaxClients > 0 {
				loadPercent = (peerCount * 100) / s.MaxClients
				if loadPercent > 100 {
					loadPercent = 100
				}
			}
		}
	} else {
		// Fallback: try to connect to WireGuard port
		healthy = h.checkPort(s.EndpointIP, s.Port)
	}

	status := "active"
	if !healthy {
		status = "unhealthy"
		loadPercent = 100
	}

	if err := h.serverRepo.UpdateHealth(ctx, s.ID, loadPercent, status, peerCount); err != nil {
		log.Printf("Health check: failed to update server %s health: %v", s.Name, err)
	}
}

func (h *HealthChecker) checkPort(host string, port int) bool {
	if host == "" {
		return false
	}
	addr := net.JoinHostPort(host, "")
	conn, err := net.DialTimeout("udp", addr, 3*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetServerStats returns current peer stats for a server
func (h *HealthChecker) GetServerStats(ctx context.Context, serverID uuid.UUID) (*wireguard.PeerStats, error) {
	if h.wgClient == nil {
		return nil, nil
	}
	device, err := h.wgClient.Device("wg0")
	if err != nil {
		return nil, err
	}

	// In a real implementation, you'd filter peers by server
	// For now, return aggregate stats
	var totalRx, totalTx int64
	var latestHandshake time.Time
	for _, peer := range device.Peers {
		totalRx += peer.ReceiveBytes
		totalTx += peer.TransmitBytes
		if !peer.LastHandshakeTime.IsZero() && (latestHandshake.IsZero() || peer.LastHandshakeTime.After(latestHandshake)) {
			latestHandshake = peer.LastHandshakeTime
		}
	}

	return &wireguard.PeerStats{
		PeerCount:       len(device.Peers),
		TotalBytesRx:    totalRx,
		TotalBytesTx:    totalTx,
		LatestHandshake: latestHandshake,
	}, nil
}