-- CreateCoinTransactionTransfer вставляет запись о переводе монет между сотрудниками.
-- $1 - id отправителя, $2 - id получателя, $3 - сумма перевода.
-- name: CreateCoinTransactionTransfer :exec
INSERT INTO coin_transactions (transaction_type, from_employee_id, to_employee_id, amount)
VALUES ('transfer', $1, $2, $3);

------------------------------------------------------------
-- CreateCoinTransactionPurchase вставляет запись о покупке мерча.
-- $1 - id сотрудника (покупателя), $2 - id мерча, $3 - сумма (обычно равна цене товара).
-- name: CreateCoinTransactionPurchase :exec
INSERT INTO coin_transactions (transaction_type, from_employee_id, merch_id, amount)
VALUES ('purchase', $1, $2, $3);
