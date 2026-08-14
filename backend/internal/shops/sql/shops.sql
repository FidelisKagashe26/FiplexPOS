-- name: CreateShop :one
INSERT INTO shops (name, address, owner_id, is_active)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetShopByID :one
SELECT * FROM shops WHERE id = $1 AND is_active = true LIMIT 1;

-- name: ListShopsByOwner :many
SELECT * FROM shops WHERE owner_id = $1 AND is_active = true ORDER BY created_at DESC;

-- name: ListAllShops :many
SELECT * FROM shops ORDER BY created_at DESC;

-- name: UpdateShop :one
UPDATE shops
SET name = COALESCE(sqlc.narg('name'), name),
    address = COALESCE(sqlc.narg('address'), address),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    updated_at = now()
WHERE id = $1
RETURNING *;
