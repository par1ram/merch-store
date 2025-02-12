-- GetEmployeeByUsername возвращает сотрудника по username.
-- name: GetEmployeeByUsername :one
SELECT 
  id,
  username,
  password_hash,
  coins,
  created_at
FROM employees
WHERE username = $1;

------------------------------------------------------------
-- CreateEmployee создаёт нового сотрудника с указанным username и password_hash.
-- При создании coins устанавливается значение по умолчанию (1000).
-- name: CreateEmployee :one
INSERT INTO employees (username, password_hash)
VALUES ($1, $2)
RETURNING id, username, coins, password_hash;

------------------------------------------------------------
-- UpdateEmployeeCoins обновляет баланс сотрудника.
-- Параметр $2 может быть положительным (начисление) или отрицательным (списание).
-- Здесь добавлена проверка, чтобы новый баланс не стал отрицательным.
-- name: UpdateEmployeeCoins :exec
UPDATE employees
SET coins = coins + $2
WHERE id = $1 AND coins + $2 >= 0;

------------------------------------------------------------
-- GetEmployeeByID возвращает данные сотрудника по его идентификатору.
-- name: GetEmployeeByID :one
SELECT 
  id,
  username,
  coins,
  created_at
FROM employees
WHERE id = $1;


------------------------------------------------------------
-- Получение баланса по ID сотрудника.
-- name: GetCoinsByID :one
SELECT coins FROM employees WHERE id=$1;
