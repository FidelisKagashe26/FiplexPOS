package repository

// Hand-written repository methods that complement the sqlc-generated Queries.
// These live outside rbac.sql.go so a future `sqlc generate` will not overwrite
// them. They are defined on the same *Queries receiver, so they share its db
// handle and transaction (WithTx) behaviour.

import (
	"context"

	"github.com/google/uuid"
)

// UpsertPermission inserts a permission by its unique name, or updates its
// module/description if it already exists, returning the permission id.
// Used by the startup seeder so the Go catalog stays the source of truth.
func (q *Queries) UpsertPermission(ctx context.Context, name, module string, description *string) (uuid.UUID, error) {
	const query = `
INSERT INTO permissions (name, module, description)
VALUES ($1, $2, $3)
ON CONFLICT (name) DO UPDATE
SET module = EXCLUDED.module,
    description = EXCLUDED.description
RETURNING id`
	var id uuid.UUID
	err := q.db.QueryRow(ctx, query, name, module, description).Scan(&id)
	return id, err
}

// ListRolePermissionNames returns the flat set of permission names granted to a role.
func (q *Queries) ListRolePermissionNames(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	const query = `
SELECT p.name
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
WHERE rp.role_id = $1`
	rows, err := q.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

// ListRolePermissionIDs returns the permission ids granted to a role.
func (q *Queries) ListRolePermissionIDs(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	const query = `SELECT permission_id FROM role_permissions WHERE role_id = $1`
	rows, err := q.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := []uuid.UUID{}
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ClearRolePermissions removes every permission currently attached to a role.
// Combined with AssignPermissionToRole this lets a handler replace a role's
// permission matrix atomically inside a transaction (via WithTx).
func (q *Queries) ClearRolePermissions(ctx context.Context, roleID uuid.UUID) error {
	_, err := q.db.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID)
	return err
}

// GetUserPermissionNamesInShop resolves the flat set of permission names a user
// holds in a given shop (across their assigned role for that shop).
func (q *Queries) GetUserPermissionNamesInShop(ctx context.Context, userID, shopID uuid.UUID) ([]string, error) {
	const query = `
SELECT DISTINCT p.name
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
JOIN user_shop_roles ur ON ur.role_id = rp.role_id
JOIN roles r ON r.id = ur.role_id
WHERE ur.user_id = $1 AND ur.shop_id = $2 AND r.is_active = true`
	rows, err := q.db.Query(ctx, query, userID, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

// GetUserPrimaryShopID returns the shop a user belongs to when no explicit
// X-Shop-Id is supplied. If the user belongs to several shops this returns the
// most recently created assignment; callers should send X-Shop-Id to disambiguate.
func (q *Queries) GetUserPrimaryShopID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	const query = `
SELECT ur.shop_id
FROM user_shop_roles ur
JOIN shops s ON s.id = ur.shop_id
WHERE ur.user_id = $1 AND s.is_active = true
ORDER BY s.created_at DESC
LIMIT 1`
	var shopID uuid.UUID
	err := q.db.QueryRow(ctx, query, userID).Scan(&shopID)
	return shopID, err
}

// IsUserInShop reports whether a user has any role assignment in a shop.
func (q *Queries) IsUserInShop(ctx context.Context, userID, shopID uuid.UUID) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM user_shop_roles WHERE user_id = $1 AND shop_id = $2)`
	var exists bool
	err := q.db.QueryRow(ctx, query, userID, shopID).Scan(&exists)
	return exists, err
}
