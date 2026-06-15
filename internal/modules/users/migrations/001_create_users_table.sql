CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    photo VARCHAR(500),
    photo_thumbnail VARCHAR(500),
    username VARCHAR(150) UNIQUE NOT NULL,
    email VARCHAR(254) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL DEFAULT '',
    is_superadmin BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_staff BOOLEAN NOT NULL DEFAULT false,
    is_verified BOOLEAN NOT NULL DEFAULT false,
    password VARCHAR(255) NOT NULL,
    password_changed_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    settings JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    updated_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT username_not_empty CHECK (username <> ''),
    CONSTRAINT email_not_empty CHECK (email <> '')
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
