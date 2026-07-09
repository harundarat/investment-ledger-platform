-- +goose Up
CREATE TYPE account_type AS ENUM (
  'asset',
  'liability',
  'equity',
  'revenue',
  'expense'
);

CREATE TABLE IF NOT EXISTS accounts (
  id UUID PRIMARY KEY,
  code INT NOT NULL,
  name VARCHAR(255) NOT NULL,
  type account_type NOT NULL,
  user_id UUID REFERENCES users(id),
  currency VARCHAR(3) NOT NULL
);

CREATE TYPE journal_entry_type AS ENUM (
  'topup',
  'buy',
  'withdrawal',
  'sell',
  'fee'
);

CREATE TABLE IF NOT EXISTS journal_entries (
  id UUID PRIMARY KEY,
  entry_type journal_entry_type NOT NULL,
  description TEXT,
  idempotency_key TEXT UNIQUE NOT NULL,
  occurred_at TIMESTAMPTZ NOT NULL
);

CREATE TYPE journal_line_direction AS ENUM (
  'debit',
  'credit'
);

CREATE TABLE IF NOT EXISTS journal_lines (
  id UUID PRIMARY KEY,
  journal_entry_id UUID NOT NULL REFERENCES journal_entries(id),
  account_id UUID NOT NULL REFERENCES accounts(id),
  direction journal_line_direction NOT NULL,
  amount BIGINT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS journal_lines;
DROP TABLE IF EXISTS journal_entries;
DROP TABLE IF EXISTS accounts;
DROP TYPE IF EXISTS account_type;
DROP TYPE IF EXISTS journal_line_direction;
DROP TYPE IF EXISTS journal_entry_type;
DROP TYPE IF EXISTS account_type;
