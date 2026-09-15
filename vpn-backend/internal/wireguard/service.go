package wireguard

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/vpn/backend/internal/server"
	"github.com/vpn/backend/internal/subscription"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	WGInterface = "wg0"
	WGPort      = 51820
)

type Service struct {
	peerRepo   *Repository
	serverRepo *server.Repository
	subRepo    *subscription.Repository
	wgClient   *wgctrl.Client
	serverIP   string
	mockMode   bool
}

func NewService(
	peerRepo *Repository,
	serverRepo *server.Repository,
	subRepo *subscription.Repository,
	serverIP string,
) (*Service, error) {
	var client *wgctrl.Client
	var mockMode bool

	if runtime.GOOS == "windows" {
		mockMode = true
	} else {
		var err error
		client, err = wgctrl.New()
		if err != nil {
			mockMode = true
		}
	}

	return &Service{
		peerRepo:   peerRepo,
		serverRepo: serverRepo,
		subRepo:    subRepo,
		wgClient:   client,
		serverIP:   serverIP,
		mockMode:   mockMode,
	}, nil
}

func (s *Service) GenerateConfig(ctx context.Context, userID, serverID uuid.UUID) (*PeerConfig, error) {
	sub, err := s.subRepo.GetActiveByUserID(ctx, userID)
	if err != nil || sub == nil || sub.Status != "active" {
		return nil, fmt.Errorf("no active subscription")
	}

	srv, err := s.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("server not found")
	}

	existing, err := s.peerRepo.GetByUserAndServer(ctx, userID, serverID)
	if err == nil && existing != nil {
		return s.buildConfig(existing, srv), nil
	}

	privateKey, err := generatePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("generate private key: %w", err)
	}

	publicKey := privateKey.PublicKey()

	presharedKey, err := generatePresharedKey()
	if err != nil {
		return nil, fmt.Errorf("generate preshared key: %w", err)
	}

	peer := &Peer{
		ID:                 uuid.New(),
		UserID:             userID,
		ServerID:           serverID,
		PrivateKey:         privateKey.String(),
		PublicKey:          publicKey.String(),
		PresharedKey:       presharedKey.String(),
		AllowedIPs:         []string{"0.0.0.0/0", "::/0"},
		DNSServers:         []string{"1.1.1.1", "1.0.0.1"},
		EndpointIP:         srv.EndpointIP,
		PersistentKeepalive: 25,
		Status:             "active",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.peerRepo.Create(ctx, peer); err != nil {
		return nil, fmt.Errorf("save peer: %w", err)
	}

	if !s.mockMode {
		if err := s.addPeerToInterface(srv, peer); err != nil {
			return nil, fmt.Errorf("add peer to interface: %w", err)
		}
	}

	return s.buildConfig(peer, srv), nil
}

func (s *Service) buildConfig(peer *Peer, srv *server.Server) *PeerConfig {
	cfg := &PeerConfig{}
	cfg.Interface.PrivateKey = peer.PrivateKey
	cfg.Interface.Address = []string{"10.0.0.2/32", "fd00::2/128"}
	cfg.Interface.DNS = peer.DNSServers

	cfg.Peer.PublicKey = srv.PublicKey
	cfg.Peer.PresharedKey = peer.PresharedKey
	cfg.Peer.AllowedIPs = peer.AllowedIPs
	cfg.Peer.Endpoint = fmt.Sprintf("%s:%d", srv.EndpointIP, srv.Port)
	cfg.Peer.PersistentKeepalive = peer.PersistentKeepalive

	return cfg
}

func (s *Service) addPeerToInterface(srv *server.Server, peer *Peer) error {
	if s.mockMode {
		return nil
	}

	pubKey, err := wgtypes.ParseKey(peer.PublicKey)
	if err != nil {
		return err
	}

	psk, err := wgtypes.ParseKey(peer.PresharedKey)
	if err != nil {
		return err
	}

	allowedIPs := []net.IPNet{}
	for _, ip := range peer.AllowedIPs {
		_, ipNet, _ := net.ParseCIDR(ip)
		if ipNet != nil {
			allowedIPs = append(allowedIPs, *ipNet)
		}
	}

	peerConfig := wgtypes.PeerConfig{
		PublicKey:         pubKey,
		PresharedKey:      &psk,
		AllowedIPs:        allowedIPs,
		PersistentKeepalive: &peer.PersistentKeepalive,
	}

	return s.wgClient.ConfigureDevice(WGInterface, wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerConfig},
	})
}

func generatePrivateKey() (*wgtypes.Key, error) {
	var key wgtypes.Key
	if _, err := rand.Read(key[:]); err != nil {
		return nil, err
	}
	return &key, nil
}

func generatePresharedKey() (*wgtypes.Key, error) {
	var key wgtypes.Key
	if _, err := rand.Read(key[:]); err != nil {
		return nil, err
	}
	return &key, nil
}