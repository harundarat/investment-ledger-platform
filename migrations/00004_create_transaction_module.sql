-- +goose Up
CREATE TYPE order_side AS ENUM (
  'buy',
  'sell'
);

CREATE TYPE order_status AS ENUM (
  'pending',
  'settled',
  'failed'
);

CREATE TABLE IF NOT EXISTS orders (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  product_id UUID NOT NULL REFERENCES products(id),
  side order_side NOT NULL,
  amount_idr BIGINT NOT NULL,
  units BIGINT NOT NULL,
  price_used BIGINT NOT NULL,
  status order_status NOT NULL,
  journal_entry_id UUID NOT NULL REFERENCES journal_entries(id)
);

CREATE TYPE cash_transaction_direction AS ENUM (
  'in',
  'out'
);

CREATE TYPE cash_transaction_status AS ENUM (
  'pending',
  'success',
  'failed'
);

CREATE TABLE IF NOT EXISTS cash_transactions (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  direction cash_transaction_direction NOT NULL,
  amount BIGINT NOT NULL,
  status cash_transaction_status NOT NULL,
  bank_reference TEXT NOT NULL,
  journal_entry_id UUID NOT NULL REFERENCES journal_entries(id)
);

-- +goose Down
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS cash_transactions;
DROP TYPE IF EXISTS order_side;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS cash_transaction_direction;
DROP TYPE IF EXISTS cash_transaction_status;
