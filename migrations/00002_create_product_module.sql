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

INSERT INTO products (id, code, type, name) VALUES
  ('00000000-0000-0000-0000-000000001001', 1001, 'mutual_fund', 'BNI-AM Indeks IDX30'),
  ('00000000-0000-0000-0000-000000001002', 1002, 'mutual_fund', 'Schroder Dana Prestasi Plus'),
  ('00000000-0000-0000-0000-000000002001', 2001, 'stock', 'Bank Central Asia Tbk'),
  ('00000000-0000-0000-0000-000000002002', 2002, 'stock', 'Telkom Indonesia Tbk'),
  ('00000000-0000-0000-0000-000000003001', 3001, 'bond', 'Obligasi Negara FR0100'),
  ('00000000-0000-0000-0000-000000003002', 3002, 'bond', 'Obligasi Negara FR0103');

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
