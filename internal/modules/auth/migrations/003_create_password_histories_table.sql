-- Migration: Create users_password_histories table
-- Menyimpan riwayat hash password untuk mencegah reuse

CREATE TABLE IF NOT EXISTS users_password_histories (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT        NOT NULL,
    password_hash VARCHAR(255)  NOT NULL,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_histories_user_id ON users_password_histories(user_id);