ALTER TABLE users ADD COLUMN is_active BOOLEAN DEFAULT TRUE;

-- Index for faster filtering since we will always query by is_active = true
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);