# VPN Project

A subscription-based VPN mobile application (Android-first, iOS later) with WireGuard protocol.

## Architecture

- **vpn-android**: Kotlin modular Android app
  - `app` - UI layer (auth screens, home, server selection, account)
  - `core-vpn` - VPN tunnel service (VpnService + com.wireguard.android)
  - `core-network` - API client (Retrofit + Moshi)
  - `core-data` - Encrypted local storage (Android Keystore + DataStore)

- **vpn-backend**: Go backend API
  - `cmd/api` - HTTP server entry point
  - `internal/auth` - JWT auth (signup, login, refresh)
  - `internal/user` - User management
  - `internal/server` - VPN server metadata
  - `internal/wireguard` - Peer config generation (wgctrl-go)
  - `internal/subscription` - Subscription management
  - `pkg/config` - Configuration
  - `pkg/jwt` - JWT token handling
  - `migrations` - Database migrations

- **vpn-infra**: Terraform infrastructure
  - `modules/vpn-node` - Reproducible VPN node provisioning
  - `environments/dev` - Development environment (3 regions)
  - `environments/prod` - Production environment

## Phase 0 Setup Complete

### Prerequisites
- Docker & Docker Compose
- Go 1.22+
- Android Studio
- Terraform 1.5+

### Quick Start

1. **Start local infrastructure:**
```bash
cd vpn-backend
docker-compose up -d
```

2. **Run backend:**
```bash
cd vpn-backend
go run ./cmd/api
```

3. **Open Android project in Android Studio** and build/run

4. **Provision VPN nodes (requires DigitalOcean + Cloudflare):**
```bash
cd vpn-infra/environments/dev
terraform init
terraform apply
```

### API Endpoints

- `POST /api/v1/auth/signup` - Create account
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh access token
- `GET /api/v1/servers` - List VPN servers (requires auth)
- `GET /api/v1/wireguard/config/{server_id}` - Get WireGuard config (requires auth + subscription)

### Database Schema

- `users` - User accounts
- `servers` - VPN server fleet
- `subscriptions` - User subscriptions (Google Play validated)
- `wg_peers` - WireGuard peer configurations

### Security Principles

- Stateless JWT auth (horizontal scaling ready)
- Server-side subscription validation (never trust client)
- Encrypted token storage (Android Keystore)
- WireGuard peer keys generated server-side
- No traffic logging (privacy by design)