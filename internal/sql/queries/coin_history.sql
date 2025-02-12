-- GetReceivedTransfers возвращает историю переводов (монеты, полученные сотрудником).
-- Для каждого перевода возвращается сумма и имя отправителя.
-- name: GetReceivedTransfers :many
SELECT 
  ct.amount,
  e.username AS from_user
FROM coin_transactions ct
JOIN employees e ON ct.from_employee_id = e.id
WHERE ct.transaction_type = 'transfer'
  AND ct.to_employee_id = $1
ORDER BY ct.created_at DESC;

------------------------------------------------------------
-- GetSentTransfers возвращает историю исходящих переводов (монеты, отправленные сотрудником).
-- Для каждого перевода возвращается сумма и имя получателя.
-- name: GetSentTransfers :many
SELECT 
  ct.amount,
  e.username AS to_user
FROM coin_transactions ct
JOIN employees e ON ct.to_employee_id = e.id
WHERE ct.transaction_type = 'transfer'
  AND ct.from_employee_id = $1
ORDER BY ct.created_at DESC;
