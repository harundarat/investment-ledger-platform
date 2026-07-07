-- +goose Up
CREATE TYPE product_type AS ENUM (
  'mutual_fund',
  'stock',
  'bond'
);

CREATE TABLE IF NOT EXISTS products (
  id UUID PRIMARY KEY,
  code INT UNIQUE NOT NULL,
  type product_type NOT NULL,
  name VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS product_prices (
  id UUID PRIMARY KEY,
  product_id UUID NOT NULL REFERENCES products(id),
  price BIGINT NOT NULL,
  price_date TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS product_prices;
DROP TYPE IF EXISTS product_type;
