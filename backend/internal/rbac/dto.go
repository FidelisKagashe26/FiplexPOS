package rbac

import (
	"time"

	"github.com/google/uuid"
)

type CreateRoleRequest struct {
	ShopID      *uuid.UUID `json:"shop_id"` // Optional; nil means a global/system role
	Name        string     `json:"name" validate:"required,min=2,max=50"`
	Description string     `json:"description"`
}

type RoleResponse struct {
	ID          uuid.UUID  `json:"id"`
	ShopID      *uuid.UUID `json:"shop_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ListRolesResponse struct {
	Data []RoleResponse `json:"data"`
}

// UpdateRolePermissionsRequest replaces the full set of permissions on a role
// (permissions.id is a UUID). Sending the complete desired set keeps the
// frontend matrix and the backend in sync in a single call.
type UpdateRolePermissionsRequest struct {
	PermissionIDs []uuid.UUID `json:"permission_ids" validate:"required"`
}

// AssignUserRoleRequest links a user to a role within a shop.
type AssignUserRoleRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	ShopID uuid.UUID `json:"shop_id,omitempty" validate:"omitempty"`
	RoleID uuid.UUID `json:"role_id" validate:"required"`
}

// PermissionResponse is a single catalog permission, for the frontend matrix.
type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Module      string    `json:"module"`
	Description string    `json:"description"`
}

// MyPermissionsResponse is what the frontend consumes to render (show/hide) UI.
// The backend remains the sole authority; this list is advisory for the client.
type MyPermissionsResponse struct {
	Role        string     `json:"role"`
	ShopID      *uuid.UUID `json:"shop_id"`
	Permissions []string   `json:"permissions"`
}
