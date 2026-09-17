# VPN Project

A subscription-based VPN mobile application (Android-first, iOS later) with WireGuard protocol.

## Architecture

- **vpn-android**: Kotlin modular Android app
  - `app` - UI layer (auth screens, home, server selection, account, settings)
  - `core-vpn` - VPN tunnel service (VpnService + com.wireguard.android) with kill switch, DNS leak protection, split tunneling
  - `core-network` - API client (Retrofit + kotlinx.serialization), BillingClient wrapper
  - `core-data` - Encrypted local storage (Android Keystore + EncryptedSharedPreferences)

- **vpn-backend**: Go backend API
  - `cmd/api` - HTTP server entry point
  - `internal/auth` - JWT auth (signup, login, refresh)
  - `internal/user` - User management
  - `internal/server` - VPN server metadata + health checks
  - `internal/wireguard` - Peer config generation (wgctrl-go)
  - `internal/subscription` - Subscription management
  - `internal/billing` - Google Play receipt validation
  - `pkg/config` - Configuration
  - `pkg/jwt` - JWT token handling
  - `pkg/redis` - Redis client
  - `pkg/ratelimit` - Redis-backed rate limiting
  - `pkg/logging` - Privacy-compliant structured logging
  - `migrations` - Database migrations

- **vpn-infra**: Terraform infrastructure
  - `modules/vpn-node` - Reproducible VPN node provisioning (DigitalOcean + Cloudflare)
  - `environments/dev` - Development environment (3 regions: nyc1, lon1, sfo3)
  - `environments/prod` - Production environment

## Phases Complete

✅ **Phase 0** - Setup: Repos, Docker, Go, Android Studio, Terraform
✅ **Phase 1** - Backend Core: Auth, WireGuard config, Postgres schema, JWT
✅ **Phase 2** - Android MVP: Modular app, VpnService, WireGuard, Hilt DI
✅ **Phase 3** - Multi-Server: Redis health cache, background checker, Terraform nodes
✅ **Phase 4** - SaaS Billing: Google Play validation, subscription gating
✅ **Phase 5** - Hardening: Kill switch, DNS leak protection, split tunneling, rate limiting, privacy logs

## Prerequisites

- Docker & Docker Compose
- Go 1.22+
- Android Studio (for Android build)
- Terraform 1.5+ (for infrastructure)

## Quick Start

### 1. Start local infrastructure:
```bash
cd vpn-backend
docker-compose up -d
```

### 2. Run database migrations:
```bash
cd vpn-backend
go run ./cmd/migrate
go run ./cmd/seed
```

### 3. Run backend API:
```bash
cd vpn-backend
go run ./cmd/api
```
API runs on `http://localhost:8080`

### 4. Build Android app:
Open `vpn-android` in Android Studio → Sync → Build → Run on emulator/device

### 5. Provision VPN nodes (requires DigitalOcean + Cloudflare):
```bash
cd vpn-infra/environments/dev
# Create terraform.tfvars with credentials
terraform init
terraform apply
```

## API Endpoints

### Auth (rate limited)
- `POST /api/v1/auth/signup` - Create account (5/15min)
- `POST /api/v1/auth/login` - Login (10/15min)
- `POST /api/v1/auth/refresh` - Refresh access token (20/15min)

### Servers (rate limited, requires auth)
- `GET /api/v1/servers` - List VPN servers with live health (30/min)
- `GET /api/v1/servers/{id}` - Get server details
- `GET /api/v1/servers/{id}/stats` - Real-time peer stats

### WireGuard (rate limited, requires auth + active subscription)
- `GET /api/v1/wireguard/config/{server_id}` - Get WireGuard config (10/hr)
- `GET /api/v1/wireguard/peers` - List user's peer configs

### Billing (rate limited, requires auth)
- `POST /api/v1/billing/validate-purchase` - Validate Google Play purchase (5/15min)
- `POST /api/v1/billing/refresh-subscription` - Re-check subscription status
- `GET /api/v1/billing/status` - Get current subscription status

### Health
- `GET /health` - Service health check

## Database Schema

- `users` - User accounts (email, password_hash)
- `servers` - VPN server fleet (region, host, public_key, status, load_percent)
- `subscriptions` - User subscriptions (plan, status, platform, purchase_token, current_period_end, auto_renew)
- `wg_peers` - WireGuard peer configurations (private_key, public_key, preshared_key, allowed_ips, endpoint)
- `rate_limits` - API rate limiting counters (in Redis, not Postgres)

## Security & Privacy Principles

- **Stateless JWT auth** - Horizontal scaling ready, no server-side session state
- **Server-side subscription validation** - Never trust client; config issuance gated by verified active subscription
- **Encrypted token storage** - Android Keystore + EncryptedSharedPreferences
- **WireGuard peer keys generated server-side** - Private keys never leave server
- **Kill switch** - Blocks all traffic if VPN tunnel drops (Android NetworkCallback)
- **DNS leak protection** - All DNS queries forced through VPN tunnel (1.1.1.1, 1.0.0.1)
- **Split tunneling** - Per-app VPN routing via UID allow/deny lists
- **Rate limiting** - Redis-backed sliding window, per-endpoint limits, fail-open
- **Privacy-compliant logging** - Automatic redaction of tokens, keys, PII, IPs; no traffic content logged

## Rate Limits (configurable in pkg/ratelimit)

| Endpoint | Requests | Window |
|----------|----------|--------|
| auth:signup | 5 | 15 min |
| auth:login | 10 | 15 min |
| auth:refresh | 20 | 15 min |
| wireguard:config | 10 | 1 hour |
| servers:list | 30 | 1 min |
| billing:validate | 5 | 15 min |

## Testing Rate Limiting
```bash
# Signup limit test (6th request returns 429)
for i in {1..6}; do 
  curl -X POST http://localhost:8080/api/v1/auth/signup \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"test$i@example.com\",\"password\":\"password123\"}"
done
```

## Project Structure
```
VPN/
├── vpn-android/          # Android app (Kotlin)
│   ├── app/              # UI layer
│   ├── core-vpn/         # VPN tunnel service
│   ├── core-network/     # API + Billing client
│   └── core-data/        # Encrypted storage
├── vpn-backend/          # Go API
│   ├── cmd/
│   │   ├── api/          # Main server
│   │   ├── migrate/      # DB migrations
│   │   └── seed/         # Test data
│   ├── internal/
│   │   ├── auth/
│   │   ├── billing/
│   │   ├── server/
│   │   ├── subscription/
│   │   ├── user/
│   │   └── wireguard/
│   ├── pkg/
│   │   ├── config/
│   │   ├── jwt/
│   │   ├── redis/
│   │   ├── ratelimit/
│   │   └── logging/
│   ├── migrations/
│   ├── docker-compose.yml
│   └── Dockerfile
└── vpn-infra/            # Terraform
    ├── modules/vpn-node/
    └── environments/
        ├── dev/
        └── prod/
```

## Next Steps (Post-Launch)

- **Phase 6** - Pre-Launch: Load testing, security audit, Play Store submission, monitoring
- **Phase 7** - iOS: NetworkExtension framework, Apple IAP validation
- **Phase 8** - Scale & Reliability: Read replicas, load balancer, auto-provisioning, DR runbook, SLOs

## License

Proprietary - All rights reserved.