-- name: CreateRole :one
INSERT INTO roles (shop_id, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1 LIMIT 1;

-- name: ListRolesByShop :many
SELECT * FROM roles WHERE shop_id = $1 OR shop_id IS NULL ORDER BY created_at DESC;

-- name: CreatePermission :one
INSERT INTO permissions (name, description, module)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListPermissions :many
SELECT * FROM permissions ORDER BY module, name;

-- name: AssignPermissionToRole :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemovePermissionFromRole :exec
DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2;

-- name: GetRolePermissions :many
SELECT p.* 
FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = $1;

-- name: AssignUserRole :exec
INSERT INTO user_shop_roles (user_id, shop_id, role_id)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, shop_id) DO UPDATE SET role_id = EXCLUDED.role_id;

-- name: GetUserRoleInShop :one
SELECT r.*
FROM roles r
JOIN user_shop_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1 AND ur.shop_id = $2;

-- name: GetUserPermissionsInShop :many
SELECT p.*
FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
JOIN user_shop_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = $1 AND ur.shop_id = $2;
