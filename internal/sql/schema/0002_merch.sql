-- +goose Up
CREATE TABLE merch (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) UNIQUE NOT NULL,
  price INTEGER NOT NULL CHECK (price > 0)
);

-- +goose Down
DROP TABLE merch;
