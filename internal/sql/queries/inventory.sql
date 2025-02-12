-- UpsertInventory добавляет товар в инвентарь сотрудника или обновляет количество,
-- если запись для данной пары (employee_id, merch_id) уже существует.
-- name: UpsertInventory :exec
INSERT INTO inventory (employee_id, merch_id, quantity)
VALUES ($1, $2, $3)
ON CONFLICT (employee_id, merch_id)
DO UPDATE SET quantity = inventory.quantity + EXCLUDED.quantity;

------------------------------------------------------------
-- GetInventoryByEmployeeID возвращает список купленных товаров сотрудника.
-- name: GetInventoryByEmployeeID :many
SELECT 
  i.employee_id,
  i.merch_id,
  m.name AS merch_name,
  i.quantity
FROM inventory i
JOIN merch m ON i.merch_id = m.id
WHERE i.employee_id = $1
ORDER BY m.name;
