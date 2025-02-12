-- GetMerchByName возвращает информацию о товаре по его названию.
-- name: GetMerchByName :one
SELECT 
  id,
  name,
  price
FROM merch
WHERE name = $1;

------------------------------------------------------------
-- ListMerch возвращает список всех товаров из мерча.
-- name: ListMerch :many
SELECT 
  id,
  name,
  price
FROM merch
ORDER BY name;
