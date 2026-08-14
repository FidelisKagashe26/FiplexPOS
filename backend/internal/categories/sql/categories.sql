-- name: CreateCategory :one
-- Create a new category, tagged with the active shop (nullable during transition).
INSERT INTO categories (name, shop_id)
VALUES ($1, sqlc.narg('shop_id'))
RETURNING *;

-- name: GetCategory :one
-- Fetch a single category by ID, scoped to the active shop.
-- Lenient scope: when no shop is active (shop_id arg NULL) all rows are visible,
-- preserving single-tenant behaviour; when a shop is active, isolation applies.
SELECT *
FROM categories
WHERE id = $1
  AND (sqlc.narg('shop_id')::uuid IS NULL OR shop_id = sqlc.narg('shop_id'))
LIMIT 1;

-- name: ListCategories :many
SELECT *
FROM categories
WHERE (sqlc.narg('shop_id')::uuid IS NULL OR shop_id = sqlc.narg('shop_id'))
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: ListCategoriesWithProducts :many
SELECT c.id, c.name, c.created_at, c.updated_at, COUNT(pc.product_id) AS product_count
FROM categories c
LEFT JOIN product_categories pc ON c.id = pc.category_id
WHERE (sqlc.narg('shop_id')::uuid IS NULL OR c.shop_id = sqlc.narg('shop_id'))
GROUP BY c.id
ORDER BY c.name ASC
LIMIT $1 OFFSET $2;


-- name: UpdateCategory :one
UPDATE categories
SET name = $2
WHERE id = $1
  AND (sqlc.narg('shop_id')::uuid IS NULL OR shop_id = sqlc.narg('shop_id'))
RETURNING *;

-- name: DeleteCategory :exec
DELETE
FROM categories
WHERE id = $1
  AND (sqlc.narg('shop_id')::uuid IS NULL OR shop_id = sqlc.narg('shop_id'));

-- name: CountCategories :one
SELECT count(*) FROM categories
WHERE (sqlc.narg('shop_id')::uuid IS NULL OR shop_id = sqlc.narg('shop_id'));

-- name: CountProductsInCategory :one
SELECT count(*) FROM product_categories WHERE category_id = $1;

-- name: ExistsCategory :one
SELECT EXISTS (
    SELECT 1
    FROM categories
    WHERE id = $1
      AND (sqlc.narg('shop_id')::uuid IS NULL OR shop_id = sqlc.narg('shop_id'))
);
