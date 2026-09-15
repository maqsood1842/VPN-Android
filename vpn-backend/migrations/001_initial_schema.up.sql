-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

-- Servers table
CREATE TABLE servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    region VARCHAR(100) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL DEFAULT 51820,
    public_key VARCHAR(255) NOT NULL,
    endpoint_ip VARCHAR(45),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    load_percent INT NOT NULL DEFAULT 0,
    max_clients INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_servers_status ON servers(status);

-- Subscriptions table
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'inactive',
    platform VARCHAR(20) NOT NULL,
    purchase_token VARCHAR(500),
    original_transaction_id VARCHAR(500),
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    auto_renew BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);

-- WireGuard peers table
CREATE TABLE wg_peers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    server_id UUID NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    private_key VARCHAR(255) NOT NULL,
    public_key VARCHAR(255) NOT NULL,
    preshared_key VARCHAR(255),
    allowed_ips CIDR[] NOT NULL DEFAULT ARRAY['0.0.0.0/0'::CIDR, '::/0'::CIDR],
    dns_servers VARCHAR(255)[] NOT NULL DEFAULT ARRAY['1.1.1.1', '1.0.0.1'],
    endpoint_ip VARCHAR(45),
    persistent_keepalive INT DEFAULT 25,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    last_handshake TIMESTAMPTZ,
    bytes_received BIGINT DEFAULT 0,
    bytes_sent BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, server_id)
);

CREATE INDEX idx_wg_peers_user_id ON wg_peers(user_id);
CREATE INDEX idx_wg_peers_server_id ON wg_peers(server_id);
CREATE INDEX idx_wg_peers_status ON wg_peers(status);

-- Rate limiting table (can be moved to Redis in production)
CREATE TABLE rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    identifier VARCHAR(255) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    count INT NOT NULL DEFAULT 1,
    window_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(identifier, endpoint, window_start)
);

CREATE INDEX idx_rate_limits_identifier ON rate_limits(identifier);