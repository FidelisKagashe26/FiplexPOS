-- name: CreateCustomer :one
INSERT INTO customers (name, phone, email, address, shop_id)
VALUES ($1, $2, $3, $4, sqlc.narg('shop_id'))
RETURNING *;

-- name: GetCustomerByID :one
SELECT * FROM customers
WHERE id = $1 AND deleted_at IS NULL
  AND (sqlc.narg('shop_id')::uuid IS NULL OR shop_id = sqlc.narg('shop_id'));

-- name: ListCustomers :many
SELECT * FROM customers
WHERE deleted_at IS NULL
  AND (sqlc.narg('shop_id')::uuid IS NULL OR shop_id = sqlc.narg('shop_id'))
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountCustomers :one
SELECT COUNT(*) FROM customers
WHERE deleted_at IS NULL
  AND (sqlc.narg('shop_id')::uuid IS NULL OR shop_id = sqlc.narg('shop_id'));

-- name: UpdateCustomer :one
UPDATE customers
SET name = $2, phone = $3, email = $4, address = $5, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
  AND (sqlc.narg('shop_id')::uuid IS NULL OR shop_id = sqlc.narg('shop_id'))
RETURNING *;

-- name: DeleteCustomer :exec
UPDATE customers SET deleted_at = NOW()
WHERE id = $1
  AND (sqlc.narg('shop_id')::uuid IS NULL OR shop_id = sqlc.narg('shop_id'));
