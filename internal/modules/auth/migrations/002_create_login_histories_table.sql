-- Migration: Create users_login_histories table
-- Menyimpan riwayat percobaan login

CREATE TABLE IF NOT EXISTS users_login_histories (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT,
    identifier     VARCHAR(255)  NOT NULL,
    ip_address     VARCHAR(45)   NOT NULL,
    user_agent     TEXT,
    status         VARCHAR(20)   NOT NULL,  -- success | failed
    failure_reason VARCHAR(255),
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_login_histories_user_id    ON users_login_histories(user_id);
CREATE INDEX IF NOT EXISTS idx_login_histories_ip         ON users_login_histories(ip_address);
CREATE INDEX IF NOT EXISTS idx_login_histories_status     ON users_login_histories(status);
CREATE INDEX IF NOT EXISTS idx_login_histories_created_at ON users_login_histories(created_at);