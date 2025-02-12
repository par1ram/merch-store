-- +goose Up
CREATE TABLE inventory (
  id SERIAL PRIMARY KEY,
  employee_id INTEGER NOT NULL REFERENCES employees(id),
  merch_id INTEGER NOT NULL REFERENCES merch(id),
  quantity INTEGER NOT NULL CHECK (quantity >= 0),
  UNIQUE (employee_id, merch_id)
);

CREATE INDEX idx_inventory_employee ON inventory(employee_id);

-- +goose Down
DROP TABLE inventory;
