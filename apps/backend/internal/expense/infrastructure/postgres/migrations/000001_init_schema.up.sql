
-- 1. Users Table
-- use TEXT for the ID so we can support simple usernames (like "Alice") 
-- for the CLI, and later support UUIDs or Emails for the web app.
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Expenses Table
CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY,
    description TEXT NOT NULL,
    total_cents BIGINT NOT NULL,
    payer_id TEXT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Splits Table
CREATE TABLE IF NOT EXISTS splits (
    id BIGSERIAL PRIMARY KEY,
    expense_id UUID NOT NULL REFERENCES expenses(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id),
    amount_cents BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_splits_expense_id ON splits(expense_id);