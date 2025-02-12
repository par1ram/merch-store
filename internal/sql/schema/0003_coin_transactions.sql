-- +goose Up
CREATE TYPE transaction_type_enum AS ENUM ('transfer', 'purchase');

CREATE TABLE coin_transactions (
  id SERIAL PRIMARY KEY,
  transaction_type transaction_type_enum NOT NULL,
  from_employee_id INTEGER NOT NULL REFERENCES employees(id),
  to_employee_id INTEGER REFERENCES employees(id),
  merch_id INTEGER REFERENCES merch(id),
  amount INTEGER NOT NULL CHECK (amount > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (
    (transaction_type = 'transfer' AND to_employee_id IS NOT NULL AND merch_id IS NULL)
    OR (transaction_type = 'purchase' AND to_employee_id IS NULL AND merch_id IS NOT NULL)
  )
);

CREATE INDEX idx_transactions_from ON coin_transactions(from_employee_id);
CREATE INDEX idx_transactions_to ON coin_transactions(to_employee_id);

-- +goose Down
DROP TYPE transaction_type_enum;
DROP TABLE coin_transactions;
