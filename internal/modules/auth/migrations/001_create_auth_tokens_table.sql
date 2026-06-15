-- Migration: Create auth_tokens table
-- Menyimpan refresh token yang aktif (outstanding tokens)

CREATE TABLE IF NOT EXISTS auth_tokens (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT        NOT NULL,
    jti          VARCHAR(255)  NOT NULL UNIQUE,
    token_type   VARCHAR(50)   NOT NULL DEFAULT 'refresh',
    device_info  TEXT,
    ip_address   VARCHAR(45),
    is_blacklist BOOLEAN       NOT NULL DEFAULT FALSE,
    expires_at   TIMESTAMPTZ   NOT NULL,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_auth_tokens_user_id      ON auth_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_jti          ON auth_tokens(jti);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_deleted_at   ON auth_tokens(deleted_at);
CREATE INDEX IF NOT EXISTS idx_auth_tokens_active       ON auth_tokens(user_id, is_blacklist, expires_at)
    WHERE deleted_at IS NULL;