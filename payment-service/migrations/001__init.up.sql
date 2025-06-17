CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL UNIQUE,
    balance DECIMAL(15, 2) NOT NULL DEFAULT 0,
);

CREATE INDEX idx_accounts_created_at ON accounts(created_at);