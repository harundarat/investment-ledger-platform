-- +goose Up
CREATE TABLE IF NOT EXISTS holdings (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  product_id UUID NOT NULL REFERENCES products(id),
  units BIGINT NOT NULL,
  avg_buy_price BIGINT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS holdings;
